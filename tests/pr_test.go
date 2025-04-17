// Tests in this file are run in the PR pipeline.
package test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"

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

const standardSolutionTerraformDir = "solutions/standard"
const fscloudExampleTerraformDir = "examples/fscloud"
const latestVersion = "7.2"

// Use existing resource group
const resourceGroup = "geretain-test-redis"

// Set up tests to only use supported BYOK regions
const regionSelectionPath = "../common-dev-assets/common-go-assets/icd-region-prefs.yaml"

// Define a struct with fields that match the structure of the YAML data
const yamlLocation = "../common-dev-assets/common-go-assets/common-permanent-resources.yaml"

var permanentResources map[string]any

var sharedInfoSvc *cloudinfo.CloudInfoService
var validICDRegions = []string{
	"eu-de",
	"us-south",
}

// TestMain will be run before any parallel tests, used to read data from yaml for use with tests
func TestMain(m *testing.M) {
	sharedInfoSvc, _ = cloudinfo.NewCloudInfoServiceFromEnv("TF_VAR_ibmcloud_api_key", cloudinfo.CloudInfoServiceOptions{})

	var err error
	permanentResources, err = common.LoadMapFromYaml(yamlLocation)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestRunStandardSolutionSchematics(t *testing.T) {
	t.Parallel()

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fmt.Sprintf("%s/*.tf", standardSolutionTerraformDir),
			fmt.Sprintf("%s/*.tf", fscloudExampleTerraformDir),
			fmt.Sprintf("%s/*.tf", "modules/fscloud"),
			fmt.Sprintf("%s/*.sh", "scripts"),
		},
		TemplateFolder:         standardSolutionTerraformDir,
		BestRegionYAMLPath:     regionSelectionPath,
		Prefix:                 "redis-st-da",
		ResourceGroup:          resourceGroup,
		DeleteWorkspaceOnFail:  false,
		WaitJobCompleteMinutes: 60,
	})

	serviceCredentialSecrets := []map[string]any{
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

	serviceCredentialNames := map[string]string{
		"admin": "Administrator",
		"user1": "Viewer",
		"user2": "Editor",
	}

	serviceCredentialNamesJSON, err := json.Marshal(serviceCredentialNames)
	if err != nil {
		log.Fatalf("Error converting to JSON: %s", err)
	}

	options.TerraformVars = []testschematic.TestSchematicTerraformVar{
		{Name: "ibmcloud_api_key", Value: options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], DataType: "string", Secure: true},
		{Name: "access_tags", Value: permanentResources["accessTags"], DataType: "list(string)"},
		{Name: "existing_kms_instance_crn", Value: permanentResources["hpcs_south_crn"], DataType: "string"},
		{Name: "existing_backup_kms_key_crn", Value: permanentResources["hpcs_south_root_key_crn"], DataType: "string"},
		{Name: "kms_endpoint_type", Value: "private", DataType: "string"},
		{Name: "resource_group_name", Value: options.Prefix, DataType: "string"},
		{Name: "redis_version", Value: "7.2", DataType: "string"}, // Always lock this test into the latest supported Redis version
		{Name: "service_credential_names", Value: string(serviceCredentialNamesJSON), DataType: "map(string)"},
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "admin_pass_secret_manager_secret_group", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass_secret_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: GetRandomAdminPassword(t), DataType: "string"},
	}
	err = options.RunSchematicTest()
	assert.Nil(t, err, "This should not have errored")
}

func TestRunStandardUpgradeSolution(t *testing.T) {
	t.Parallel()

	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       standardSolutionTerraformDir,
		BestRegionYAMLPath: regionSelectionPath,
		Prefix:             "redis-st-da-upg",
		ResourceGroup:      resourceGroup,
	})

	options.TerraformVars = map[string]any{
		"access_tags":               permanentResources["accessTags"],
		"existing_kms_instance_crn": permanentResources["hpcs_south_crn"],
		"kms_endpoint_type":         "public",
		"provider_visibility":       "public",
		"resource_group_name":       options.Prefix,
	}

	output, err := options.RunTestUpgrade()
	if !options.UpgradeTestSkipped {
		assert.Nil(t, err, "This should not have errored")
		assert.NotNil(t, output, "Expected some output")
	}
}

func TestPlanValidation(t *testing.T) {
	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  standardSolutionTerraformDir,
		Prefix:        "validate-plan",
		ResourceGroup: resourceGroup,
		Region:        "us-south", // skip VPC region picker
	})
	options.TestSetup()
	options.TerraformOptions.NoColor = true
	options.TerraformOptions.Logger = logger.Discard
	options.TerraformOptions.Vars = map[string]any{
		"prefix":              options.Prefix,
		"region":              "us-south",
		"redis_version":       "7.2",
		"provider_visibility": "public",
		"resource_group_name": options.Prefix,
	}

	// Test the DA when using an existing KMS instance
	var standardSolutionWithExistingKms = map[string]any{
		"access_tags":               permanentResources["accessTags"],
		"existing_kms_instance_crn": permanentResources["hpcs_south_crn"],
	}

	// Test the DA when using IBM owned encryption key
	var standardSolutionWithUseIbmOwnedEncKey = map[string]any{
		"use_ibm_owned_encryption_key": true,
	}

	// Create a map of the variables
	tfVarsMap := map[string]map[string]any{
		"standardSolutionWithExistingKms":       standardSolutionWithExistingKms,
		"standardSolutionWithUseIbmOwnedEncKey": standardSolutionWithUseIbmOwnedEncKey,
	}

	_, initErr := terraform.InitE(t, options.TerraformOptions)
	if assert.Nil(t, initErr, "This should not have errored") {
		// Iterate over the slice of maps
		for name, tfVars := range tfVarsMap {
			t.Run(name, func(t *testing.T) {
				// Iterate over the keys and values in each map
				for key, value := range tfVars {
					options.TerraformOptions.Vars[key] = value
				}
				output, err := terraform.PlanE(t, options.TerraformOptions)
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
	prefix := fmt.Sprintf("redis-t-%s", strings.ToLower(random.UniqueId()))
	realTerraformDir := ".."
	tempTerraformDir, _ := files.CopyTerraformFolderToTemp(realTerraformDir, fmt.Sprintf(prefix+"-%s", strings.ToLower(random.UniqueId())))

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(validICDRegions))))
	if err != nil {
		log.Fatalf("Failed to generate a secure random index: %v", err)
	}
	region := validICDRegions[index.Int64()]

	// Verify ibmcloud_api_key variable is set
	checkVariable := "TF_VAR_ibmcloud_api_key"
	val, present := os.LookupEnv(checkVariable)
	require.True(t, present, checkVariable+" environment variable not set")
	require.NotEqual(t, "", val, checkVariable+" environment variable is empty")

	logger.Log(t, "Tempdir: ", tempTerraformDir)
	existingTerraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTerraformDir + "/examples/basic",
		Vars: map[string]any{
			"prefix":            prefix,
			"region":            region,
			"redis_version":     latestVersion,
			"service_endpoints": "public-and-private",
		},
		// Set Upgrade to true to ensure latest version of providers and modules are used by terratest.
		// This is the same as setting the -upgrade=true flag with terraform.
		Upgrade: true,
	})

	terraform.WorkspaceSelectOrNew(t, existingTerraformOptions, prefix)
	_, existErr := terraform.InitAndApplyE(t, existingTerraformOptions)
	if existErr != nil {
		assert.True(t, existErr == nil, "Init and Apply of temp existing resource failed")
	} else {
		logger.Log(t, " existing_redis_instance_crn: ", terraform.Output(t, existingTerraformOptions, "redis_crn"))
		options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
			Testing: t,
			TarIncludePatterns: []string{
				"*.tf",
				fmt.Sprintf("%s/*.tf", standardSolutionTerraformDir),
				fmt.Sprintf("%s/*.tf", fscloudExampleTerraformDir),
				fmt.Sprintf("%s/*.tf", "modules/fscloud"),
				fmt.Sprintf("%s/*.sh", "scripts"),
			},
			TemplateFolder:         standardSolutionTerraformDir,
			BestRegionYAMLPath:     regionSelectionPath,
			Prefix:                 "redis-da",
			ResourceGroup:          resourceGroup,
			DeleteWorkspaceOnFail:  false,
			WaitJobCompleteMinutes: 60,
		})

		options.TerraformVars = []testschematic.TestSchematicTerraformVar{
			{Name: "ibmcloud_api_key", Value: options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], DataType: "string", Secure: true},
			{Name: "existing_redis_instance_crn", Value: terraform.Output(t, existingTerraformOptions, "redis_crn"), DataType: "string"},
			{Name: "resource_group_name", Value: fmt.Sprintf("%s-resource-group", prefix), DataType: "string"},
			{Name: "region", Value: region, DataType: "string"},
			{Name: "use_existing_resource_group", Value: true, DataType: "bool"},
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
		terraform.Destroy(t, existingTerraformOptions)
		terraform.WorkspaceDelete(t, existingTerraformOptions, prefix)
		logger.Log(t, "END: Destroy (existing resources)")
	}
}

func GetRandomAdminPassword(t *testing.T) string {
	// Generate a 15 char long random string for the admin_pass
	randomBytes := make([]byte, 13)
	_, randErr := rand.Read(randomBytes)
	require.Nil(t, randErr) // do not proceed if we can't gen a random password
	randomPass := "A1" + base64.URLEncoding.EncodeToString(randomBytes)[:13]
	return randomPass
}
