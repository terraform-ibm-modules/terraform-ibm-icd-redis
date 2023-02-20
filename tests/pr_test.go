// Tests in this file are run in the PR pipeline
package test

import (
	"os"
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

// TestMain will be run before any parallel tests, used to read data from yaml for use with tests
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestRunner(t *testing.T) {
	t.Parallel()
	t.Run("Tests", func(t *testing.T) {
		t.Run("testRunDefaultExample", testRunDefaultExample)
		t.Run("testRunRedisAutoScaleExample", testRunRedisAutoScaleExample)
		t.Run("testRunCompleteExample", testRunCompleteExample)
		t.Run("testRunRedisFSCloudExample", testRunRedisFSCloudExample)
		t.Run("testRunUpgradeExample", testRunUpgradeExample)
	})
}
func testRunDefaultExample(t *testing.T) {
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

func testRunRedisAutoScaleExample(t *testing.T) {
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

func testRunRedisFSCloudExample(t *testing.T) {
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
func testRunCompleteExample(t *testing.T) {
	t.Parallel()
	versions := []string{"5", "6"}
	for _, version := range versions {
		t.Run(version, func(t *testing.T) { testRunComplete(t, version) })
	}
}

func testRunUpgradeExample(t *testing.T) {
	// TODO: Remove this line after the first merge to primary branch is complete to enable upgrade test
	t.Skip("Skipping upgrade test until initial code is in primary branch")
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
