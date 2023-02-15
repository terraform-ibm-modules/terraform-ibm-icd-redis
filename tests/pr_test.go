// Tests in this file are run in the PR pipeline
package test

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testhelper"
	"gopkg.in/yaml.v2"
)

// Use existing resource group
const resourceGroup = "geretain-test-redis"
const defaultExampleTerraformDir = "examples/default"
const completeExampleTerraformDir = "examples/complete"
const fsCloudTerraformDir = "examples/fscloud"
const autoscalingExampleTerraformDir = "examples/redis-auto-scaling"

// Restricting due to limited availability of BYOK in certain regions
const regionSelectionPath = "../common-dev-assets/common-go-assets/icd-region-prefs.yaml"

// Define a struct with fields that match the structure of the YAML data
const yamlLocation = "../common-dev-assets/common-go-assets/common-permanent-resources.yaml"

type Config struct {
	SmGuid   string `yaml:"secretsManagerGuid"`
	SmRegion string `yaml:"secretsManagerRegion"`
}

var smGuid string
var smRegion string

// TestMain will be run before any parallel tests, used to read data from yaml for use with tests
func TestMain(m *testing.M) {
	// Read the YAML file contents
	data, err := os.ReadFile(yamlLocation)
	if err != nil {
		log.Fatal(err)
	}
	// Create a struct to hold the YAML data
	var config Config
	// Unmarshal the YAML data into the struct
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	// Parse the SM guid and region from data
	smGuid = config.SmGuid
	smRegion = config.SmRegion
	os.Exit(m.Run())
}

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
			"redis_version":               version,
			"existing_sm_instance_guid":   smGuid,
			"existing_sm_instance_region": smRegion,
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
	// TODO: Remove this line after the first merge to primary branch is complete to enable upgrade test
	t.Skip("Skipping upgrade test until initial code is in primary branch")
	t.Parallel()

	options := testhelper.TestOptionsDefaultWithVars(&testhelper.TestOptions{
		Testing:            t,
		TerraformDir:       completeExampleTerraformDir,
		Prefix:             "redis-template-upg",
		ResourceGroup:      resourceGroup,
		BestRegionYAMLPath: regionSelectionPath,
		TerraformVars: map[string]interface{}{
			"existing_sm_instance_guid":   smGuid,
			"existing_sm_instance_region": smRegion,
		},
	})

	output, err := options.RunTestUpgrade()
	if !options.UpgradeTestSkipped {
		assert.Nil(t, err, "This should not have errored")
		assert.NotNil(t, output, "Expected some output")
	}
}
