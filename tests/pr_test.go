// Tests in this file are run in the PR pipeline
package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testhelper"
)

// Use existing resource group
const resourceGroup = "geretain-test-redis"
const defaultExampleTerraformDir = "examples/default"
const completeExampleTerraformDir = "examples/complete"
const fsCloudTerraformDir = "examples/fscloud"
const autoscalingExampleTerraformDir = "examples/redis-auto-scaling"

// Restricting due to limited availability of BYOK in certain regions
const regionSelectionPath = "../common-dev-assets/common-go-assets/icd-region-prefs.yaml"

func TestRunDefaultExample(t *testing.T) {
	t.Parallel()

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:       t,
		TerraformDir:  defaultExampleTerraformDir,
		Prefix:        "redis",
		ResourceGroup: resourceGroup,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

func TestRunRedisAutoScaleExample(t *testing.T) {
	t.Parallel()

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       autoscalingExampleTerraformDir,
		Prefix:             "redis",
		ResourceGroup:      resourceGroup,
		BestRegionYAMLPath: regionSelectionPath,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

func TestRunRedisFSCloudExample(t *testing.T) {
	t.Parallel()

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       fsCloudTerraformDir,
		Prefix:             "redis",
		ResourceGroup:      resourceGroup,
		BestRegionYAMLPath: regionSelectionPath,
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}

func testRunComplete(t *testing.T, version string) {
	t.Parallel()
	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       completeExampleTerraformDir,
		Prefix:             "redis-" + version,
		ResourceGroup:      resourceGroup,
		BestRegionYAMLPath: regionSelectionPath,

		TerraformVars: map[string]interface{}{
			"redis_version": version,
		},
	})

	output, err := options.RunTestConsistency()
	assert.Nil(t, err, "This should not have errored")
	assert.NotNil(t, output, "Expected some output")
}
func TestRunCompleteExample(t *testing.T) {
	t.Parallel()
	versions := []string{"5", "6"}
	for _, version := range versions {
		t.Run(version, func(t *testing.T) { testRunComplete(t, version) })
	}
}

func TestRunUpgradeExample(t *testing.T) {
	t.Parallel()
	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       completeExampleTerraformDir,
		Prefix:             "redis-template-upg",
		ResourceGroup:      resourceGroup,
		BestRegionYAMLPath: regionSelectionPath,
	})

	output, err := options.RunTestUpgrade()
	if !options.UpgradeTestSkipped {
		assert.Nil(t, err, "This should not have errored")
		assert.NotNil(t, output, "Expected some output")
	}
}
