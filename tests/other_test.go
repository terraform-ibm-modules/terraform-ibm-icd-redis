// Tests in this file are NOT run in the PR pipeline. They are run in the continuous testing pipeline along with the ones in pr_test.go
package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testhelper"
)

func testPlanICDVersions(t *testing.T, version string) {
	t.Parallel()

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:      t,
		TerraformDir: "examples/basic",
		TerraformVars: map[string]interface{}{
			"redis_version": version,
		},
		CloudInfoService: sharedInfoSvc,
	})
	output, err := options.RunTestPlan()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

func TestPlanICDVersions(t *testing.T) {
	t.Parallel()

	// This test will run a terraform plan on available stable versions of redis
	versions, _ := sharedInfoSvc.GetAvailableIcdVersions("redis")
	for _, version := range versions {
		t.Run(version, func(t *testing.T) { testPlanICDVersions(t, version) })
	}
}

// Test the DA when using IBM owned encryption keys
func TestRunStandardSolutionIBMKeys(t *testing.T) {
	t.Parallel()

	options := testhelper.TestOptionsDefault(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  fullyConfigurableSolutionTerraformDir,
		Region:        "us-south",
		Prefix:        "redis-key",
		ResourceGroup: resourceGroup,
	})

	options.TerraformVars = map[string]interface{}{
		"redis_version":                "7.2",
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
