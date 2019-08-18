package grid

import (
	"testing"

	"gotest.tools/assert"
)

func TestReadConfig(t *testing.T) {
	testConfig, err := readConfig("grid_test_config")
	t.Log(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertStaticConfig := StaticConfig{"Grid Intertie", 20, 10}
	assertDynamicConfig := DynamicConfig{20}
	assertConfig := Config{Static: assertStaticConfig, Dynamic: assertDynamicConfig}
	assert.Assert(t, testConfig == assertConfig)
}
