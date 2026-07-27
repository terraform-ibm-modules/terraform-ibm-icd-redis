// Tests in this file are run in the PR pipeline
package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/IBM/cloud-databases-go-sdk/clouddatabasesv5"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/google/uuid"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/cloudinfo"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/common"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testhelper"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testschematic"
)

const fullyConfigurableSolutionTerraformDir = "solutions/fully-configurable"

const icdType = "redis"
const icdShortType = "redis"

// Use existing resource group
const resourceGroup = "geretain-test-redis"

// Restricting due to limited availability of BYOK in certain regions
const regionSelectionPath = "../common-dev-assets/common-go-assets/icd-region-prefs.yaml"

// Define a struct with fields that match the structure of the YAML data
const yamlLocation = "../common-dev-assets/common-go-assets/common-permanent-resources.yaml"

var permanentResources map[string]interface{}

var sharedInfoSvc *cloudinfo.CloudInfoService
var validICDRegions = []string{
	"eu-de",
	"us-south",
}

func GetLatestAndOldestVersions(icdAvailableVersions []string) (string, string) {

	if len(icdAvailableVersions) == 0 {
		log.Fatal("No available ICD versions found")
	}

	sort.Slice(icdAvailableVersions, func(i, j int) bool {
		partsI := strings.Split(icdAvailableVersions[i], ".")
		partsJ := strings.Split(icdAvailableVersions[j], ".")

		majorI, _ := strconv.Atoi(partsI[0])
		majorJ, _ := strconv.Atoi(partsJ[0])

		if majorI != majorJ {
			return majorI < majorJ
		}

		minorI := 0
		minorJ := 0

		if len(partsI) >= 2 {
			minorI, _ = strconv.Atoi(partsI[1])
		}
		if len(partsJ) >= 2 {
			minorJ, _ = strconv.Atoi(partsJ[1])
		}
		return minorI < minorJ
	})

	fmt.Println("version list is ", icdAvailableVersions)
	latestVersion := icdAvailableVersions[len(icdAvailableVersions)-1]
	oldestVersion := icdAvailableVersions[0]

	return latestVersion, oldestVersion

}

func GetRegionVersions(region string) (string, string) {

	icdRegion := region
	if region == "ca-mon" {
		icdRegion = "ca-tor"
	}

	cloudInfoSvc, err := cloudinfo.NewCloudInfoServiceFromEnv("TF_VAR_ibmcloud_api_key", cloudinfo.CloudInfoServiceOptions{
		IcdRegion: icdRegion,
	})

	if err != nil {
		log.Fatal(err)
	}

	icdAvailableVersions, err := cloudInfoSvc.GetAvailableIcdVersions(icdType)

	if err != nil {
		log.Fatal(err)
	}

	return GetLatestAndOldestVersions(icdAvailableVersions)
}

func GetVersionsGen2(region string, plan string) (string, string) {

	cloudInfoSvc, err := cloudinfo.NewCloudInfoServiceFromEnv("TF_VAR_ibmcloud_api_key", cloudinfo.CloudInfoServiceOptions{})

	if err != nil {
		log.Fatal(err)
	}

	icdAvailableVersions, err := cloudInfoSvc.GetAvailableIcdVersionsGen2("databases-for-redis", plan, region) // this function takes service, plan and region as arguments in this specific order

	if err != nil {
		log.Fatal(err)
	}

	return GetLatestAndOldestVersions(icdAvailableVersions)
}

// GetLatestGen2BackupCRN uses the IBM Cloud Databases SDK to fetch the most recent
// completed, restorable backup for the given gen2 deployment CRN.
// The ibm_database_backups Terraform data source does not support gen2 deployments,
// so this helper is used to retrieve the backup CRN programmatically.
//
// Gen2 backups are served by the regional ICD API endpoint. The region is extracted
// from the deployment CRN (field index 5). If no backup is available yet (the instance
// was recently provisioned), the function polls every 10 minutes for up to 90 minutes.
func GetLatestGen2BackupCRN(deploymentCRN string) string {
	apiKey := os.Getenv("TF_VAR_ibmcloud_api_key")
	if apiKey == "" {
		log.Fatal("TF_VAR_ibmcloud_api_key environment variable not set")
	}

	// Extract region from CRN: crn:v1:cname:ctype:service:REGION:scope:instance:rtype:resource
	parts := strings.Split(deploymentCRN, ":")
	if len(parts) < 6 || parts[5] == "" {
		log.Fatalf("Cannot extract region from deployment CRN: %s", deploymentCRN)
	}
	region := parts[5]

	// Build the region-specific ICD API URL
	// Format: https://api.{region}.databases.cloud.ibm.com/v5/ibm
	regionalURL, err := clouddatabasesv5.ConstructServiceURL(map[string]string{
		"region":   region,
		"platform": "ibm",
	})
	if err != nil {
		log.Fatalf("Error constructing ICD API URL for region %s: %v", region, err)
	}

	icdClient, err := clouddatabasesv5.NewCloudDatabasesV5(&clouddatabasesv5.CloudDatabasesV5Options{
		URL: regionalURL,
		Authenticator: &core.IamAuthenticator{
			ApiKey: apiKey, // pragma: allowlist secret
		},
	})
	if err != nil {
		log.Fatalf("Error creating Cloud Databases client: %v", err)
	}

	// Poll until a completed, restorable backup is found.
	// A newly provisioned gen2 instance may not have a backup for up to ~90 minutes.
	const (
		maxWait  = 90 * time.Minute
		interval = 10 * time.Minute
	)
	deadline := time.Now().Add(maxWait)

	for {
		listOpts := icdClient.NewListDeploymentBackupsOptions(deploymentCRN)
		backups, _, err := icdClient.ListDeploymentBackups(listOpts)
		if err != nil {
			fmt.Printf("Warning: listing backups for %s failed: %v — retrying in %s\n", deploymentCRN, err, interval)
		} else if backups != nil && len(backups.Backups) > 0 {
			// Find the latest completed and restorable backup
			var latestBackupCRN string
			var latestTime time.Time
			for _, b := range backups.Backups {
				if b.ID == nil || b.Status == nil || b.IsRestorable == nil {
					continue
				}
				if *b.Status != "completed" || !*b.IsRestorable {
					continue
				}
				if b.CreatedAt != nil && time.Time(*b.CreatedAt).After(latestTime) {
					latestTime = time.Time(*b.CreatedAt)
					latestBackupCRN = *b.ID
				}
			}
			if latestBackupCRN != "" {
				fmt.Printf("Latest gen2 backup CRN for %s: %s\n", deploymentCRN, latestBackupCRN)
				return latestBackupCRN
			}
		}

		if time.Now().After(deadline) {
			log.Fatalf("No completed, restorable backup found for deployment %s after %s", deploymentCRN, maxWait)
		}
		fmt.Printf("No backup available yet for %s, waiting %s (deadline: %s)...\n", deploymentCRN, interval, deadline.Format(time.RFC3339))
		time.Sleep(interval)
	}
}

func TestRunBasicGen2Example(t *testing.T) {
	t.Parallel()

	latestVersion, _ := GetVersionsGen2("eu-de", "standard-gen2")
	fmt.Println("Latest version is ", latestVersion)

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       "examples/basic",
		Prefix:             "redis-gen2",
		BestRegionYAMLPath: regionSelectionPath,
		ResourceGroup:      resourceGroup,
		TerraformVars: map[string]interface{}{ // Limited gen2 to eu-de
			"region":            "eu-de",
			"plan":              "standard-gen2",
			"redis_version":     latestVersion,
			"service_endpoints": "private",
		},
		CloudInfoService: sharedInfoSvc,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

// TestMain will be run before any parallel tests, used to read data from yaml for use with tests
func TestMain(m *testing.M) {
	var err error
	sharedInfoSvc, err = cloudinfo.NewCloudInfoServiceFromEnv("TF_VAR_ibmcloud_api_key", cloudinfo.CloudInfoServiceOptions{})
	if err != nil {
		log.Fatal(err)
	}

	permanentResources, err = common.LoadMapFromYaml(yamlLocation)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

// Test the fully-configurable DA with defaults (IBM owned encryption keys)
func TestRunFullyConfigurableSolutionSchematics(t *testing.T) {
	t.Parallel()

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fullyConfigurableSolutionTerraformDir + "/*.tf",
		},
		TemplateFolder:             fullyConfigurableSolutionTerraformDir,
		BestRegionYAMLPath:         regionSelectionPath,
		Prefix:                     fmt.Sprintf("%s-fc-da", icdShortType),
		ResourceGroup:              resourceGroup,
		DeleteWorkspaceOnFail:      false,
		WaitJobCompleteMinutes:     60,
		CheckApplyResultForUpgrade: true,
	})

	uniqueResourceGroup := generateUniqueResourceGroupName(options.Prefix)

	serviceCredentialSecrets := []map[string]interface{}{
		{
			"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
			"service_credentials": []map[string]string{
				{
					"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
					"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
				},
				{
					"secret_name": fmt.Sprintf("%s-cred-writer", options.Prefix),
					"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Editor",
				},
			},
		},
	}

	serviceCredentialNames := []map[string]string{
		{
			"name":     "redis-admin",
			"role":     "Administrator",
			"endpoint": "private",
		},
	}

	region := "us-south"
	latestVersion, _ := GetRegionVersions(region)
	options.TerraformVars = []testschematic.TestSchematicTerraformVar{
		{Name: "prefix", Value: options.Prefix, DataType: "string"},
		{Name: "ibmcloud_api_key", Value: options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], DataType: "string", Secure: true},
		{Name: "access_tags", Value: permanentResources["accessTags"], DataType: "list(string)"},
		{Name: "deletion_protection", Value: false, DataType: "bool"},
		{Name: "existing_resource_group_name", Value: uniqueResourceGroup, DataType: "string"},
		{Name: "region", Value: region, DataType: "string"},
		{Name: "service_credential_names", Value: serviceCredentialNames, DataType: "list(object)"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_group", Value: fmt.Sprintf("%s-%s-admin-secrets", icdShortType, options.Prefix), DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: common.GetRandomPasswordWithPrefix(), DataType: "string"},
		{Name: "redis_version", Value: latestVersion, DataType: "string"}, // Always lock this test into the latest supported Redis version
	}

	err := sharedInfoSvc.WithNewResourceGroup(uniqueResourceGroup, func() error {
		return options.RunSchematicTest()
	})
	assert.Nil(t, err, "This should not have errored")
}

// Upgrade test the fully-configurable DA with KMS encryption (KYOK)
func TestRunFullyConfigurableWithKMSUpgradeSolution(t *testing.T) {
	t.Parallel()

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fullyConfigurableSolutionTerraformDir + "/*.tf",
		},
		TemplateFolder:             fullyConfigurableSolutionTerraformDir,
		Tags:                       []string{fmt.Sprintf("%s-fc-upg", icdShortType)},
		Prefix:                     fmt.Sprintf("%s-fc-upg", icdShortType),
		DeleteWorkspaceOnFail:      false,
		WaitJobCompleteMinutes:     120,
		CheckApplyResultForUpgrade: true,
	})

	serviceCredentialSecrets := []map[string]interface{}{
		{
			"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
			"service_credentials": []map[string]string{
				{
					"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
					"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
				},
				{
					"secret_name": fmt.Sprintf("%s-cred-writer", options.Prefix),
					"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Editor",
				},
			},
		},
	}

	resourceKeys := []map[string]string{
		{
			"name":     "admin",
			"role":     "Administrator",
			"endpoint": "private",
		},
		{
			"name":     "user1",
			"role":     "Viewer",
			"endpoint": "private",
		},
		{
			"name":     "user2",
			"role":     "Editor",
			"endpoint": "private",
		},
	}

	uniqueResourceGroup := generateUniqueResourceGroupName(options.Prefix)

	region := "us-south"
	latestVersion, _ := GetRegionVersions(region)
	options.TerraformVars = []testschematic.TestSchematicTerraformVar{
		{Name: "prefix", Value: options.Prefix, DataType: "string"},
		{Name: "ibmcloud_api_key", Value: options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], DataType: "string", Secure: true},
		{Name: "access_tags", Value: permanentResources["accessTags"], DataType: "list(string)"},
		{Name: "deletion_protection", Value: false, DataType: "bool"},
		{Name: "region", Value: region, DataType: "string"},
		{Name: "existing_resource_group_name", Value: uniqueResourceGroup, DataType: "string"},
		{Name: "service_credential_names", Value: resourceKeys, DataType: "list(object)"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_group", Value: fmt.Sprintf("%s-%s-admin-secrets", icdShortType, options.Prefix), DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: common.GetRandomPasswordWithPrefix(), DataType: "string"},
		{Name: "kms_encryption_enabled", Value: true, DataType: "bool"},
		{Name: "existing_kms_instance_crn", Value: permanentResources["hpcs_south_crn"], DataType: "string"},
		{Name: "redis_version", Value: latestVersion, DataType: "string"}, // Always lock this test into the latest supported Redis version
	}
	err := sharedInfoSvc.WithNewResourceGroup(uniqueResourceGroup, func() error {
		return options.RunSchematicUpgradeTest()
	})
	if !options.UpgradeTestSkipped {
		assert.Nil(t, err, "This should not have errored")
	}
}

func TestPlanValidation(t *testing.T) {
	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:      t,
		TerraformDir: fullyConfigurableSolutionTerraformDir,
		Prefix:       "val-plan",
		// ResourceGroup: resourceGroup,
		Region: "us-south", // skip VPC region picker
	})
	options.TestSetup()
	options.TerraformOptions.NoColor = true
	options.TerraformOptions.Logger = logger.Discard

	latestVersion, _ := GetRegionVersions("us-south")
	options.TerraformOptions.Vars = map[string]interface{}{
		"prefix":                       options.Prefix,
		"region":                       "us-south",
		"redis_version":                latestVersion,
		"provider_visibility":          "public",
		"existing_resource_group_name": resourceGroup,
	}

	// Test the DA when using an existing KMS instance
	var fullyConfigurableWithExistingKms = map[string]interface{}{
		"access_tags":               permanentResources["accessTags"],
		"existing_kms_instance_crn": permanentResources["hpcs_south_crn"],
		"kms_encryption_enabled":    true,
	}

	// Test the DA when using IBM owned encryption key
	var fullyConfigurableWithIbmOwnedKey = map[string]interface{}{
		"kms_encryption_enabled": false,
	}

	// Test the DA when using IBM owned encryption keys
	var fullyConfigurableWithIbmOwnedBackupKey = map[string]interface{}{
		"use_default_backup_encryption_key": false,
		"kms_encryption_enabled":            false,
	}

	// Create a map of the variables
	tfVarsMap := map[string]map[string]interface{}{
		"fullyConfigurableWithExistingKms":       fullyConfigurableWithExistingKms,
		"fullyConfigurableWithIbmOwnedKey":       fullyConfigurableWithIbmOwnedKey,
		"fullyConfigurableWithIbmOwnedBackupKey": fullyConfigurableWithIbmOwnedBackupKey,
	}

	_, initErr := terraform.InitContextE(t, context.Background(), options.TerraformOptions)
	if assert.Nil(t, initErr, "This should not have errored") {
		// Iterate over the slice of maps
		for name, tfVars := range tfVarsMap {
			t.Run(name, func(t *testing.T) {
				// Iterate over the keys and values in each map
				for key, value := range tfVars {
					options.TerraformOptions.Vars[key] = value
				}
				output, err := terraform.PlanContextE(t, context.Background(), options.TerraformOptions)
				assert.Nil(t, err, "This should not have errored")
				assert.NotNil(t, output, "Expected some output")
				// Delete the keys from the map
				for key := range tfVars {
					delete(options.TerraformOptions.Vars, key)
				}
			})
		}
	}
}

func TestRunExistingInstance(t *testing.T) {
	t.Parallel()
	prefix := fmt.Sprintf("%s-t-%s", icdShortType, strings.ToLower(random.UniqueID()))
	realTerraformDir := ".."
	tempTerraformDir, _ := files.CopyTerraformFolderToTemp(realTerraformDir, fmt.Sprintf(prefix+"-%s", strings.ToLower(random.UniqueID())))

	// Verify ibmcloud_api_key variable is set
	checkVariable := "TF_VAR_ibmcloud_api_key"
	val, present := os.LookupEnv(checkVariable)
	require.True(t, present, checkVariable+" environment variable not set")
	require.NotEqual(t, "", val, checkVariable+" environment variable is empty")

	logger.Log(t, "Tempdir: ", tempTerraformDir)

	region := validICDRegions[common.CryptoIntn(len(validICDRegions))]
	_, oldestVersion := GetRegionVersions(region)
	existingTerraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTerraformDir + "/examples/basic",
		Vars: map[string]interface{}{
			"prefix":            prefix,
			"region":            region,
			"redis_version":     oldestVersion,
			"service_endpoints": "public-and-private",
		},
		// Set Upgrade to true to ensure latest version of providers and modules are used by terratest.
		// This is the same as setting the -upgrade=true flag with terraform.
		Upgrade: true,
	})

	terraform.WorkspaceSelectOrNewContext(t, context.Background(), existingTerraformOptions, prefix)
	_, existErr := terraform.InitAndApplyContextE(t, context.Background(), existingTerraformOptions)
	if existErr != nil {
		assert.True(t, existErr == nil, "Init and Apply of temp existing resource failed")
	} else {
		logger.Log(t, " existing_redis_instance_crn: ", terraform.OutputContext(t, context.Background(), existingTerraformOptions, "redis_crn"))
		options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
			Testing: t,
			TarIncludePatterns: []string{
				"*.tf",
				fullyConfigurableSolutionTerraformDir + "/*.tf",
			},
			TemplateFolder:         fullyConfigurableSolutionTerraformDir,
			BestRegionYAMLPath:     regionSelectionPath,
			Prefix:                 fmt.Sprintf("%s-ex", icdShortType),
			ResourceGroup:          resourceGroup,
			DeleteWorkspaceOnFail:  false,
			WaitJobCompleteMinutes: 60,
		})

		options.TerraformVars = []testschematic.TestSchematicTerraformVar{
			{Name: "prefix", Value: options.Prefix, DataType: "string"},
			{Name: "ibmcloud_api_key", Value: options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], DataType: "string", Secure: true},
			{Name: "existing_redis_instance_crn", Value: terraform.OutputContext(t, context.Background(), existingTerraformOptions, "redis_crn"), DataType: "string"},
			{Name: "existing_resource_group_name", Value: fmt.Sprintf("%s-resource-group", prefix), DataType: "string"},
			{Name: "deletion_protection", Value: false, DataType: "bool"},
			{Name: "region", Value: region, DataType: "string"},
			{Name: "provider_visibility", Value: "public", DataType: "string"},
		}
		err := options.RunSchematicTest()
		assert.Nil(t, err, "This should not have errored")

	}
	envVal, _ := os.LookupEnv("DO_NOT_DESTROY_ON_FAILURE")
	// Destroy the temporary existing resources if required
	if t.Failed() && strings.ToLower(envVal) == "true" {
		fmt.Println("Terratest failed. Debug the test and delete resources manually.")
	} else {
		logger.Log(t, "START: Destroy (existing resources)")
		terraform.DestroyContext(t, context.Background(), existingTerraformOptions)
		terraform.WorkspaceDeleteContext(t, context.Background(), existingTerraformOptions, prefix)
		logger.Log(t, "END: Destroy (existing resources)")
	}
}

func generateUniqueResourceGroupName(baseName string) string {
	id := uuid.New().String()[:8] // Shorten UUID for readability
	return fmt.Sprintf("%s-%s", baseName, id)
}
