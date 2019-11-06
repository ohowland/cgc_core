package feeder

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

var CONFIGPATH = "feeder_test_config.json"

func TestReadConfig(t *testing.T) {
	testConfig := Config{}
	jsonConfig, err := ioutil.ReadFile(CONFIGPATH)
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertConfig := Config{"TEST_Virtual Feeder", "Virtual Bus", 20, 19}
	assert.Assert(t, testConfig == assertConfig)
}

func TestNew(t *testing.T) {
	jsonConfig, err := ioutil.ReadFile(CONFIGPATH)
	feeder, err := New(jsonConfig, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(feeder)
	assert.Assert(t, feeder.Name() == "TEST_Virtual Feeder")
}
