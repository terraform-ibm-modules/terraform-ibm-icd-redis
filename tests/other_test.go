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

	// TODO: replace with a real Gen2 Redis instance CRN and region once a permanent Gen2 instance is available
	redisGen2Crn := "crn:v1:bluemix:public:databases-for-redis:eu-de:a/abac0df06b644a9cabc6e44f55b3880e:replace-with-real-gen2-instance-guid::"
	redisGen2Region := "eu-de"

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  "examples/backup-restore",
		Prefix:        "redis-gen2-restored",
		Region:        redisGen2Region,
		ResourceGroup: resourceGroup,
		TerraformVars: map[string]interface{}{
			"existing_database_crn": redisGen2Crn,
			"plan":                  "standard-gen2",
			"region":                redisGen2Region,
		},
		CloudInfoService: sharedInfoSvc,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}
