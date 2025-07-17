// Tests in this file are run in the PR pipeline.
package test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/cloudinfo"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/common"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testaddons"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testschematic"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"
)

const fullyConfigurableSolutionTerraformDir = "solutions/fully-configurable"
const securityEnforcedTerraformDir = "solutions/security-enforced"
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

// Test the fully-configurable DA with defaults (no KMS encryption)
func TestRunFullyConfigurableSolutionSchematics(t *testing.T) {
	t.Parallel()

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fmt.Sprintf("%s/*.tf", fullyConfigurableSolutionTerraformDir),
			fmt.Sprintf("%s/*.sh", "scripts"),
		},
		TemplateFolder:     fullyConfigurableSolutionTerraformDir,
		BestRegionYAMLPath: regionSelectionPath,
		Prefix:             "r-fc-da",
		// ResourceGroup:              resourceGroup,
		DeleteWorkspaceOnFail:      false,
		CheckApplyResultForUpgrade: true,
		WaitJobCompleteMinutes:     60,
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
		{Name: "existing_resource_group_name", Value: resourceGroup, DataType: "string"},
		{Name: "redis_version", Value: latestVersion, DataType: "string"}, // Always lock this test into the latest supported Redis version
		{Name: "service_credential_names", Value: string(serviceCredentialNamesJSON), DataType: "map(string)"},
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "admin_pass_secrets_manager_secret_group", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "prefix", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: GetRandomAdminPassword(t), DataType: "string"},
	}
	err = options.RunSchematicTest()
	assert.Nil(t, err, "This should not have errored")
}

// Test the security-enforced DA with defaults (KMS encryption enabled)
func TestRunSecurityEnforcedSolutionSchematics(t *testing.T) {
	t.Parallel()

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fmt.Sprintf("%s/*.tf", securityEnforcedTerraformDir),
			fmt.Sprintf("%s/*.tf", fullyConfigurableSolutionTerraformDir),
			fmt.Sprintf("%s/*.sh", "scripts"),
		},
		TemplateFolder:     securityEnforcedTerraformDir,
		BestRegionYAMLPath: regionSelectionPath,
		Prefix:             "r-se-da",
		// ResourceGroup:              resourceGroup,
		DeleteWorkspaceOnFail:      false,
		CheckApplyResultForUpgrade: true,
		WaitJobCompleteMinutes:     60,
	})
	fmt.Print(options)

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
		{Name: "existing_resource_group_name", Value: resourceGroup, DataType: "string"},
		{Name: "redis_version", Value: "7.2", DataType: "string"}, // Always lock this test into the latest supported Redis version
		{Name: "service_credential_names", Value: string(serviceCredentialNamesJSON), DataType: "map(string)"},
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "admin_pass_secrets_manager_secret_group", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: GetRandomAdminPassword(t), DataType: "string"},
		{Name: "prefix", Value: options.Prefix, DataType: "string"},
	}
	err = options.RunSchematicTest()
	assert.Nil(t, err, "This should not have errored")
}

func TestRunSecurityEnforcedUpgradeSolution(t *testing.T) {
	t.Parallel()

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fullyConfigurableSolutionTerraformDir + "/*.tf",
			securityEnforcedTerraformDir + "/*.tf",
			"scripts/*.sh",
		},
		TemplateFolder:     securityEnforcedTerraformDir,
		BestRegionYAMLPath: regionSelectionPath,
		Prefix:             "re-da-upg",
		// ResourceGroup:          resourceGroup,
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
		{Name: "prefix", Value: options.Prefix, DataType: "string"},
		{Name: "access_tags", Value: permanentResources["accessTags"], DataType: "list(string)"},
		{Name: "existing_kms_instance_crn", Value: permanentResources["hpcs_south_crn"], DataType: "string"},
		{Name: "existing_resource_group_name", Value: resourceGroup, DataType: "string"},
		{Name: "redis_version", Value: "7.2", DataType: "string"}, // Always lock this test into the latest supported Redis version
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "service_credential_names", Value: string(serviceCredentialNamesJSON), DataType: "map(string)"},
		{Name: "admin_pass_secrets_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: GetRandomAdminPassword(t), DataType: "string"},
		{Name: "admin_pass_secrets_manager_secret_group", Value: fmt.Sprintf("redis-%s-admin-secrets", options.Prefix), DataType: "string"},
	}

	err = options.RunSchematicUpgradeTest()
	if !options.UpgradeTestSkipped {
		assert.Nil(t, err, "This should not have errored")
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
				fmt.Sprintf("%s/*.tf", fullyConfigurableSolutionTerraformDir),
				fmt.Sprintf("%s/*.sh", "scripts"),
			},
			TemplateFolder:         fullyConfigurableSolutionTerraformDir,
			BestRegionYAMLPath:     regionSelectionPath,
			Prefix:                 "redis-da",
			ResourceGroup:          resourceGroup,
			DeleteWorkspaceOnFail:  false,
			WaitJobCompleteMinutes: 60,
		})

		options.TerraformVars = []testschematic.TestSchematicTerraformVar{
			{Name: "ibmcloud_api_key", Value: options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], DataType: "string", Secure: true},
			{Name: "existing_redis_instance_crn", Value: terraform.Output(t, existingTerraformOptions, "redis_crn"), DataType: "string"},
			{Name: "region", Value: region, DataType: "string"},
			{Name: "existing_resource_group_name", Value: fmt.Sprintf("%s-resource-group", prefix), DataType: "string"},
			{Name: "provider_visibility", Value: "public", DataType: "string"},
			{Name: "prefix", Value: options.Prefix, DataType: "string"},
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

// TestRunAddonTests runs addon tests in parallel using a matrix approach
// This can be used as an example of how to run multiple addon tests in parallel
func TestRunAddonTests(t *testing.T) {
	testCases := []testaddons.AddonTestCase{
		{
			Name:   "Redis-Default-Config",
			Prefix: "redisdef",
		},
		{
			Name:   "Redis-With-Resource-Group-And-Account-Settings",
			Prefix: "redisrgwaccs",
			Dependencies: []cloudinfo.AddonConfig{
				{
					OfferingName:   "deploy-arch-ibm-account-infra-base",
					OfferingFlavor: "resource-groups-with-account-settings",
					Enabled:        core.BoolPtr(true),
				},
			},
			SkipInfrastructureDeployment: true, // Skip infrastructure deployment for this test case
		},
		{
			Name:   "Redis-With-Resource-Group-Only",
			Prefix: "rediswrgly",
			Dependencies: []cloudinfo.AddonConfig{
				{
					OfferingName:   "deploy-arch-ibm-account-infra-base",
					OfferingFlavor: "resource-group-only",
					Enabled:        core.BoolPtr(true),
				},
			},
			SkipInfrastructureDeployment: true, // Skip infrastructure deployment for this test case
		},
		{
			Name:   "Redis-With-KMS-Disabled",
			Prefix: "redisnokm",
			Dependencies: []cloudinfo.AddonConfig{
				{
					OfferingName:   "deploy-arch-ibm-kms",
					OfferingFlavor: "fully-configurable",
					Enabled:        core.BoolPtr(false),
				},
			},
			Inputs: map[string]interface{}{
				"existing_kms_instance_crn": permanentResources["kp_us_south_root_key_crn"],
			},
			SkipInfrastructureDeployment: true, // Skip infrastructure deployment for this test case
		},
		{
			Name:   "Redis-With-Secret-Manager-Disabled",
			Prefix: "redisnosm",
			Dependencies: []cloudinfo.AddonConfig{
				{
					OfferingName:   "deploy-arch-ibm-secrets-manager",
					OfferingFlavor: "fully-configurable",
					Enabled:        core.BoolPtr(false),
				},
			},
			SkipInfrastructureDeployment: true, // Skip infrastructure deployment for this test case
		},
		{
			Name:   "Redis-With-KMS-And-Secret-Manager-Disabled",
			Prefix: "redisnokmsm",
			Dependencies: []cloudinfo.AddonConfig{
				{
					OfferingName:   "deploy-arch-ibm-kms",
					OfferingFlavor: "fully-configurable",
					Enabled:        core.BoolPtr(false),
				},
				{
					OfferingName:   "deploy-arch-ibm-secrets-manager",
					OfferingFlavor: "fully-configurable",
					Enabled:        core.BoolPtr(true),
				},
			},
			Inputs: map[string]interface{}{
				"existing_kms_instance_crn": permanentResources["kp_us_south_root_key_crn"],
			},
			SkipInfrastructureDeployment: true, // Skip infrastructure deployment for this test case
		},
		{
			Name:   "Redis-With-All-Optional-Services-Disabled",
			Prefix: "redisallno",
			Dependencies: []cloudinfo.AddonConfig{
				{
					OfferingName:   "deploy-arch-ibm-secrets-manager",
					OfferingFlavor: "fully-configurable",
					Enabled:        core.BoolPtr(false),
				},
				{
					OfferingName:   "deploy-arch-ibm-kms",
					OfferingFlavor: "fully-configurable",
					Enabled:        core.BoolPtr(false),
				},
			},
			Inputs: map[string]interface{}{
				"existing_kms_instance_crn": permanentResources["kp_us_south_root_key_crn"],
			},
			SkipInfrastructureDeployment: true,
		},
	}

	// Define common options that apply to all test cases
	baseOptions := testaddons.TestAddonsOptionsDefault(&testaddons.TestAddonOptions{
		Testing:       t,
		Prefix:        "Redis-matrix", // Test cases will override with their own prefixes
		ResourceGroup: resourceGroup,
		//SkipLocalChangeCheck: true, // Skip local change check for addon tests
	})

	matrix := testaddons.AddonTestMatrix{
		TestCases:   testCases,
		BaseOptions: baseOptions,
		BaseSetupFunc: func(baseOptions *testaddons.TestAddonOptions, testCase testaddons.AddonTestCase) *testaddons.TestAddonOptions {
			// The framework automatically handles prefix assignment from testCase.Prefix
			// You can add any custom logic here if needed
			return baseOptions
		},
		AddonConfigFunc: func(options *testaddons.TestAddonOptions, testCase testaddons.AddonTestCase) cloudinfo.AddonConfig {
			return cloudinfo.NewAddonConfigTerraform(
				options.Prefix,
				"deploy-arch-ibm-icd-redis",
				"fully-configurable",
				map[string]interface{}{
					"prefix": options.Prefix,
					"region": "us-south",
				},
			)
		},
	}
	baseOptions.RunAddonTestMatrix(matrix)
}
