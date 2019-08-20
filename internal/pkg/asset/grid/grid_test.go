package grid

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

func TestReadConfig(t *testing.T) {
	testConfig := Config{}
	jsonConfig, err := ioutil.ReadFile("grid_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertStaticConfig := StaticConfig{"Grid Intertie", 20, 10}
	assertDynamicConfig := DynamicConfig{20}
	assertConfig := Config{Static: assertStaticConfig, Dynamic: assertDynamicConfig}
	assert.Assert(t, testConfig == assertConfig)
}
