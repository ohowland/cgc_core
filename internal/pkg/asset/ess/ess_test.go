package ess

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

func TestReadConfig(t *testing.T) {
	testConfig := Config{}
	jsonConfig, err := ioutil.ReadFile("ess_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	assertConfig := Config{"TEST_Virtual ESS", "Virtual Bus", 20, 10, 50}
	assert.Assert(t, testConfig == assertConfig)
}
