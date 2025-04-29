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
	"unicode"

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
		Prefix:             "redis-fc-da",
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
		{Name: "access_tags", Value: permanentResources["accessTags"], DataType: "list(string)"},
		{Name: "existing_resource_group_name", Value: resourceGroup, DataType: "string"},
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
		Prefix:             "redis-se-da",
		// ResourceGroup:          resourceGroup,
		DeleteWorkspaceOnFail:  false,
		WaitJobCompleteMinutes: 60,
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
		{Name: "existing_backup_kms_key_crn", Value: permanentResources["hpcs_south_root_key_crn"], DataType: "string"},
		{Name: "existing_resource_group_name", Value: resourceGroup, DataType: "string"},
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

	options := testschematic.TestSchematicOptionsDefault(&testschematic.TestSchematicOptions{
		Testing: t,
		TarIncludePatterns: []string{
			"*.tf",
			fmt.Sprintf("%s/*.tf", fullyConfigurableSolutionTerraformDir),
			fmt.Sprintf("%s/*.sh", "scripts"),
		},
		TemplateFolder:     fullyConfigurableSolutionTerraformDir,
		BestRegionYAMLPath: regionSelectionPath,
		Prefix:             "redis-fc-da-upg",
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
		{Name: "access_tags", Value: permanentResources["accessTags"], DataType: "list(string)"},
		{Name: "existing_resource_group_name", Value: resourceGroup, DataType: "string"},
		{Name: "redis_version", Value: "7.2", DataType: "string"}, // Always lock this test into the latest supported Redis version
		{Name: "service_credential_names", Value: string(serviceCredentialNamesJSON), DataType: "map(string)"},
		{Name: "existing_secrets_manager_instance_crn", Value: permanentResources["secretsManagerCRN"], DataType: "string"},
		{Name: "service_credential_secrets", Value: serviceCredentialSecrets, DataType: "list(object)"},
		{Name: "admin_pass_secret_manager_secret_group", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass_secret_manager_secret_name", Value: options.Prefix, DataType: "string"},
		{Name: "admin_pass", Value: GetRandomAdminPassword(t), DataType: "string"},
	}

	err = options.RunSchematicUpgradeTest()
	if !options.UpgradeTestSkipped {
		assert.Nil(t, err, "This should not have errored")
	}
}

func cleanString(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r == '\n' || r == '\t' || unicode.IsControl(r) {
			builder.WriteRune(' ') // replace with a space
		} else {
			builder.WriteRune(r)
		}
	}
	// Replace multiple spaces with a single space and trim
	return strings.Join(strings.Fields(builder.String()), " ")
}

func TestPlanPositiveValidation(t *testing.T) {
	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  securityEnforcedTerraformDir,
		Prefix:        "validate-plan",
		ResourceGroup: resourceGroup,
		Region:        "us-south", // skip VPC region picker
	})
	options.TestSetup()
	options.TerraformOptions.NoColor = true
	options.TerraformOptions.Logger = logger.Discard
	options.TerraformOptions.Vars = map[string]any{
		"prefix":                       options.Prefix,
		"region":                       "us-south",
		"redis_version":                "7.2",
		"existing_resource_group_name": resourceGroup,
	}

	// Test the DA when using an existing KMS instance
	var securityEnforcedRequiredKms = map[string]any{
		"use_ibm_owned_encryption_key": false,
		"existing_kms_instance_crn":    permanentResources["hpcs_south_root_key_crn"],
	}

	var securityEnforcedIbmEncryptionKeyFalse1 = map[string]any{
		"existing_kms_instance_crn":    permanentResources["hpcs_south_crn"],
		"use_ibm_owned_encryption_key": false,
	}

	var securityEnforcedIbmEncryptionKeyFalse2 = map[string]any{
		"existing_kms_key_crn":         permanentResources["hpcs_south_root_key_crn"],
		"use_ibm_owned_encryption_key": false,
	}

	var securityEnforcedIbmEncryptionKeyFalse3 = map[string]any{
		"existing_kms_instance_crn":    permanentResources["hpcs_south_crn"],
		"existing_backup_kms_key_crn":  permanentResources["hpcs_south_root_key_crn"],
		"use_ibm_owned_encryption_key": false,
	}

	var securityEnforcedRegionNotMatched = map[string]any{
		"existing_redis_instance_crn": permanentResources["redisCrn"],
		"region":                      "us-south",
	}

	var securityEnforcedBackupCRN = map[string]any{
		"backup_crn":           "crn:v1:bluemix:public:databases-for-enterprisedb:us-south:a/abac0df06b644a9cabc6e44f55b3880e:3d3077b9-fcc0-427f-a877-d7af3830ca5e:backup:",
		"existing_kms_key_crn": permanentResources["hpcs_south_root_key_crn"],
	}

	var securityEnforcedServiceCredentialsRoleIsMissing = map[string]any{
		"use_ibm_owned_encryption_key":          false,
		"existing_kms_instance_crn":             permanentResources["hpcs_south_crn"],
		"existing_secrets_manager_instance_crn": permanentResources["secretsManagerCRN"],
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
	}

	var securityEnforcedExistingSMCrnRequired = map[string]any{
		"use_ibm_owned_encryption_key":          false,
		"existing_kms_instance_crn":             permanentResources["hpcs_south_crn"],
		"existing_secrets_manager_instance_crn": permanentResources["secretsManagerCRN"],
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
	}

	var securityEnforcedSMsecretGroupRequired = map[string]any{
		"use_ibm_owned_encryption_key":           false,
		"existing_kms_instance_crn":              permanentResources["hpcs_south_crn"],
		"existing_secrets_manager_instance_crn":  permanentResources["secretsManagerCRN"],
		"admin_pass_secret_manager_secret_group": "test-group",
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
	}

	var securityEnforcedSMsecretNameRequired = map[string]any{
		"use_ibm_owned_encryption_key":          false,
		"existing_kms_instance_crn":             permanentResources["hpcs_south_crn"],
		"existing_secrets_manager_instance_crn": permanentResources["secretsManagerCRN"],
		"admin_pass_secret_manager_secret_name": "secret-name",
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
	}

	// Create a map of the variables
	tfVarsMap := map[string]map[string]any{
		"securityEnforcedRequiredKms":                     securityEnforcedRequiredKms,
		"securityEnforcedIbmEncryptionKeyFalse1":          securityEnforcedIbmEncryptionKeyFalse1,
		"securityEnforcedIbmEncryptionKeyFalse2":          securityEnforcedIbmEncryptionKeyFalse2,
		"securityEnforcedIbmEncryptionKeyFalse3":          securityEnforcedIbmEncryptionKeyFalse3,
		"securityEnforcedRegionNotMatched":                securityEnforcedRegionNotMatched,
		"securityEnforcedBackupCRN":                       securityEnforcedBackupCRN,
		"securityEnforcedServiceCredentialsRoleIsMissing": securityEnforcedServiceCredentialsRoleIsMissing,
		"securityEnforcedExistingSMCrnRequired":           securityEnforcedExistingSMCrnRequired,
		"securityEnforcedSMsecretGroupRequired":           securityEnforcedSMsecretGroupRequired,
		"securityEnforcedSMsecretNameRequired":            securityEnforcedSMsecretNameRequired,
	}

	_, initErr := terraform.InitE(t, options.TerraformOptions)
	if assert.Nil(t, initErr, "This should not have errored") {
		// Iterate over the slice of maps
		for name, tfVars := range tfVarsMap {
			t.Run(name, func(t *testing.T) {

				// Create a temporary JSON file that will be used for input variables.
				tmpFile, err := os.CreateTemp("", "tfvars-*.json")
				if err != nil {
					panic(err)
				}

				// Clean up after execution
				defer func() {
					if err := os.Remove(tmpFile.Name()); err != nil {
						fmt.Printf("failed to remove temp file: %v\n", err)
					}
				}()

				// Marshal map to JSON
				jsonBytes, err := json.Marshal(tfVars)
				if err != nil {
					panic(err)
				}

				// Write content to temp file
				_, err = tmpFile.Write(jsonBytes)
				if err != nil {
					panic(err)
				}

				if err := tmpFile.Close(); err != nil {
					panic(err)
				}

				// set input '-var-file=' to temp file
				options.TerraformOptions.VarFiles = []string{tmpFile.Name()}

				// run terraform plan
				output, err := terraform.PlanE(t, options.TerraformOptions)

				// check of errors
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

func TestPlanNegativeValidation(t *testing.T) {
	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  securityEnforcedTerraformDir,
		Prefix:        "validate-plan",
		ResourceGroup: resourceGroup,
		Region:        "us-south", // skip VPC region picker
	})
	options.TestSetup()
	options.TerraformOptions.NoColor = true
	options.TerraformOptions.Logger = logger.Discard
	options.TerraformOptions.Vars = map[string]any{
		"prefix":                       options.Prefix,
		"region":                       "us-south",
		"redis_version":                "7.2",
		"existing_resource_group_name": resourceGroup,
	}

	// Test the DA when using an existing KMS instance
	var securityEnforcedRequiredKms = map[string]any{
		"use_ibm_owned_encryption_key": false,
		"error_message":                "When 'kms_encryption_enabled' is true and 'use_ibm_owned_encryption_key' is false, you must provide either 'existing_kms_instance_crn' (to create a new key) or 'existing_kms_key_crn' (to use an existing key).",
	}

	var securityEnforcedIbmEncryptionKeyFalse1 = map[string]any{
		"existing_kms_instance_crn":    permanentResources["hpcs_south_crn"],
		"use_ibm_owned_encryption_key": true,
		"error_message":                "When 'kms_encryption_enabled' is true and setting values for 'existing_kms_instance_crn', 'existing_kms_key_crn' or 'existing_backup_kms_key_crn', the 'use_ibm_owned_encryption_key' input must be set to false.",
	}

	var securityEnforcedIbmEncryptionKeyFalse2 = map[string]any{
		"existing_kms_key_crn":         permanentResources["hpcs_south_root_key_crn"],
		"use_ibm_owned_encryption_key": true,
		"error_message":                "When 'kms_encryption_enabled' is true and setting values for 'existing_kms_instance_crn', 'existing_kms_key_crn' or 'existing_backup_kms_key_crn', the 'use_ibm_owned_encryption_key' input must be set to false.",
	}

	var securityEnforcedIbmEncryptionKeyFalse3 = map[string]any{
		"existing_backup_kms_key_crn":  permanentResources["hpcs_south_root_key_crn"],
		"use_ibm_owned_encryption_key": true,
		"error_message":                "When 'kms_encryption_enabled' is true and setting values for 'existing_kms_instance_crn', 'existing_kms_key_crn' or 'existing_backup_kms_key_crn', the 'use_ibm_owned_encryption_key' input must be set to false.",
	}

	var securityEnforcedRegionNotMatched = map[string]any{
		"existing_redis_instance_crn": permanentResources["hpcs_south_crn"],
		"region":                      "us-east",
		"error_message":               "The region detected in the 'existing_redis_instance_crn' value must match the value of the 'region' input variable when passing an existing instance.",
	}

	var securityEnforcedBackupCRN = map[string]any{
		"backup_crn":           "crn:v1:bluemix:public:databases-for-enterprisedb:us-south:a/abac0df06b644a9cabc6e44f55b3880e:3d3077b9-fcc0-427f-a877-d7af3830ca5e:back:",
		"error_message":        "backup_crn must be null OR starts with 'crn:' and contains ':backup:'",
		"existing_kms_key_crn": permanentResources["hpcs_south_root_key_crn"],
	}

	// Test the DA when using an existing KMS instance
	var securityEnforcedServiceCredentialsRoleIsMissing = map[string]any{
		"use_ibm_owned_encryption_key": false,
		"existing_kms_instance_crn":    permanentResources["hpcs_south_crn"],
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::roe:Viewer",
					},
				},
			},
		},
		"error_message": "service_credentials_source_service_role_crn must be a serviceRole CRN. See https://cloud.ibm.com/iam/roles",
	}

	var securityEnforcedExistingSMCrnRequired = map[string]any{
		"use_ibm_owned_encryption_key": false,
		"existing_kms_instance_crn":    permanentResources["hpcs_south_crn"],
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
		"error_message": "`existing_secrets_manager_instance_crn` is required when adding service credentials to a secrets manager secret.",
	}

	var securityEnforcedSMsecretGroupRequired = map[string]any{
		"use_ibm_owned_encryption_key":           false,
		"existing_kms_instance_crn":              permanentResources["hpcs_south_crn"],
		"existing_secrets_manager_instance_crn":  permanentResources["secretsManagerCRN"],
		"admin_pass_secret_manager_secret_group": nil,
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
		"error_message": "`admin_pass_secret_manager_secret_group` is required when `existing_secrets_manager_instance_crn` is set.",
	}

	var securityEnforcedSMsecretNameRequired = map[string]any{
		"use_ibm_owned_encryption_key":          false,
		"existing_kms_instance_crn":             permanentResources["hpcs_south_crn"],
		"existing_secrets_manager_instance_crn": permanentResources["secretsManagerCRN"],
		"admin_pass_secret_manager_secret_name": nil,
		"service_credential_secrets": []map[string]any{
			{
				"secret_group_name": fmt.Sprintf("%s-secret-group", options.Prefix),
				"service_credentials": []map[string]string{
					{
						"secret_name": fmt.Sprintf("%s-cred-reader", options.Prefix),
						"service_credentials_source_service_role_crn": "crn:v1:bluemix:public:iam::::role:Viewer",
					},
				},
			},
		},
		"error_message": "`admin_pass_secret_manager_secret_name` is required when `existing_secrets_manager_instance_crn` is set.",
	}

	// Create a map of the variables
	tfVarsMap := map[string]map[string]any{
		"securityEnforcedRequiredKms":                     securityEnforcedRequiredKms,
		"securityEnforcedIbmEncryptionKeyFalse1":          securityEnforcedIbmEncryptionKeyFalse1,
		"securityEnforcedIbmEncryptionKeyFalse2":          securityEnforcedIbmEncryptionKeyFalse2,
		"securityEnforcedIbmEncryptionKeyFalse3":          securityEnforcedIbmEncryptionKeyFalse3,
		"securityEnforcedRegionNotMatched":                securityEnforcedRegionNotMatched,
		"securityEnforcedBackupCRN":                       securityEnforcedBackupCRN,
		"securityEnforcedServiceCredentialsRoleIsMissing": securityEnforcedServiceCredentialsRoleIsMissing,
		"securityEnforcedExistingSMCrnRequired":           securityEnforcedExistingSMCrnRequired,
		"securityEnforcedSMsecretGroupRequired":           securityEnforcedSMsecretGroupRequired,
		"securityEnforcedSMsecretNameRequired":            securityEnforcedSMsecretNameRequired,
	}

	_, initErr := terraform.InitE(t, options.TerraformOptions)
	if assert.Nil(t, initErr, "This should not have errored") {
		// Iterate over the slice of maps
		for name, tfVars := range tfVarsMap {
			t.Run(name, func(t *testing.T) {

				// Extract the error message from the map and remove it, ensuring it's not passed as an input variable.
				error_message := tfVars["error_message"]
				delete(tfVars, "error_message")

				// Create a temporary JSON file that will be used for input variables.
				tmpFile, err := os.CreateTemp("", "tfvars-*.json")
				if err != nil {
					panic(err)
				}

				// Clean up after execution
				defer func() {
					if err := os.Remove(tmpFile.Name()); err != nil {
						fmt.Printf("failed to remove temp file: %v\n", err)
					}
				}()

				// Marshal map to JSON
				jsonBytes, err := json.Marshal(tfVars)
				if err != nil {
					panic(err)
				}

				// Write content to temp file
				_, err = tmpFile.Write(jsonBytes)
				if err != nil {
					panic(err)
				}

				if err := tmpFile.Close(); err != nil {
					panic(err)
				}

				// set input '-var-file=' to temp file
				options.TerraformOptions.VarFiles = []string{tmpFile.Name()}

				// run terraform plan
				_, err = terraform.PlanE(t, options.TerraformOptions)

				// check of errors
				assert.NotNil(t, err, `'%s' Error should be thrown.`, error_message)
				if err != nil {
					assert.Contains(t, cleanString(err.Error()), cleanString(error_message.(string)), `The error message is not correct. It should be '%s'. Instead it is '%s'`, error_message, cleanString(err.Error()))
				}
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
