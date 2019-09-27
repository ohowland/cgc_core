package feeder

/*
func TestReadConfig(t *testing.T) {
	testConfig := Config{}
	jsonConfig, err := ioutil.ReadFile("ipc30c3_test_comm.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertStaticConfig := StaticConfig{"ESS", 20, 10, 0.6}
	assertDynamicConfig := DynamicConfig{}
	assertConfig := Config{Static: assertStaticConfig, Dynamic: assertDynamicConfig}
	assert.Assert(t, testConfig == assertConfig)
}
*/
