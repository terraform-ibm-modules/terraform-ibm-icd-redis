// Tests in this file are NOT run in the PR pipeline. They are run in the continuous testing pipeline along with the ones in pr_test.go
package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testhelper"
)

// Test the DA when using IBM owned encryption keys
func TestRunStandardSolutionIBMKeys(t *testing.T) {
	t.Parallel()

	region := "us-south"

	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  fullyConfigurableSolutionTerraformDir,
		Region:        region,
		Prefix:        "redis-key",
		ResourceGroup: resourceGroup,
	})

	latestVersion, _ := GetRegionVersions(region)
	options.TerraformVars = map[string]interface{}{
		"redis_version":                latestVersion,
		"region":                       region,
		"provider_visibility":          "public",
		"existing_resource_group_name": resourceGroup,
		"prefix":                       options.Prefix,
		"deletion_protection":          false,
	}

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

func TestRunRestoredDBExample(t *testing.T) {
	t.Parallel()

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  "examples/backup-restore",
		Prefix:        "redis-restored",
		Region:        fmt.Sprint(permanentResources["redisRegion"]),
		ResourceGroup: resourceGroup,
		TerraformVars: map[string]interface{}{
			"existing_database_crn": permanentResources["redisCrn"],
		},
		CloudInfoService: sharedInfoSvc,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

func TestRunRestoredDBGen2Example(t *testing.T) {
	t.Parallel()

	const gen2Region = "eu-de"
	const gen2DeploymentCRN = "crn:v1:bluemix:public:databases-for-redis:eu-de:a/abac0df06b644a9cabc6e44f55b3880e:d238b078-fd92-4dfa-b027-267282644410::"

	latestVersion, _ := GetVersionsGen2(gen2Region, "standard-gen2")
	backupCRN := GetLatestGen2BackupCRN(gen2DeploymentCRN)

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  "examples/backup-restore",
		Prefix:        "redis-gen2-restored",
		Region:        gen2Region,
		ResourceGroup: resourceGroup,
		TerraformVars: map[string]interface{}{
			"plan":          "standard-gen2",
			"redis_version": latestVersion,
			"backup_crn":    backupCRN,
		},
		CloudInfoService: sharedInfoSvc,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}
