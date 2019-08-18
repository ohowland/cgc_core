package sel1547

import (
	"encoding/json"
	"testing"

	"github.com/ohowland/cgc/internal/pkg/comm"
	"gotest.tools/assert"
)

func TestReadCommConfig(t *testing.T) {
	testComm, err := readCommConfig("sel1547_test_comm")
	if err != nil {
		t.Fatal(err)
	}
	assertPoller := comm.PollerConfig{
		IPAddr:       "192.168.0.100",
		Port:         "5020",
		SlaveID:      0,
		Timeout:      100,
		PollRate:     1000,
		EnableLogger: false,
	}
	assert.Assert(t, testComm.TargetConfig == assertPoller)

	assertRegister := []comm.Register{
		{
			Name:         "test1",
			Address:      0,
			DataType:     comm.DataType("u16"),
			FunctionCode: 3,
			AccessType:   comm.Access("read-write"),
			Endianness:   comm.Endian("big")},
		{
			Name:         "test2",
			Address:      1,
			DataType:     comm.DataType("u16"),
			FunctionCode: 3,
			AccessType:   comm.Access("read-write"),
			Endianness:   comm.Endian("big")},
		{
			Name:         "test3",
			Address:      2,
			DataType:     comm.DataType("f32"),
			FunctionCode: 3,
			AccessType:   comm.Access("read-write"),
			Endianness:   comm.Endian("little")},
	}

	assert.Assert(t, len(testComm.Registers) == len(assertRegister[:]))
	for i := range testComm.Registers {
		assert.Assert(t, testComm.Registers[i] == assertRegister[i])
	}
}

func TestMarshal(t *testing.T) {

	response := make(map[string]interface{})
	response["Kw"] = 10
	response["Kvar"] = 20
	response["Synchronized"] = false

	testJSON, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	testStatus := Status{}
	err = json.Unmarshal(testJSON, &testStatus)
	if err != nil {
		t.Fatal(err)
	}

	assertStatus := Status{Kw: 10, Kvar: 20, Synchronized: false}
	assert.Assert(t, testStatus == assertStatus)

}

/*
func TestNewAsset(t *testing.T) {
	testConfig := NewAsset("sel1547_test.json")

	assertConfig := Sel1547{
		sel1547Status{}
		sel1547Control{}

		sel1547StaticConfig{"Grid Intertie", 20, 10}
		assert.Assert(t, testConfig == assertConfig)
}
*/
