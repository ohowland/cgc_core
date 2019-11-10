package ess

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"gotest.tools/assert"
)

type DummyDevice struct{}

// randMachineStatus returns a closure for random MachineStatus
func randMachineStatus() func() MachineStatus {
	status := MachineStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false, false}
	return func() MachineStatus {
		return status
	}
}

var assertedStatus = randMachineStatus()

func (d DummyDevice) ReadDeviceStatus() (MachineStatus, error) {
	time.Sleep(100 * time.Millisecond)
	return assertedStatus(), nil
}

func (d DummyDevice) WriteDeviceControl(MachineControl) error {
	time.Sleep(100 * time.Millisecond)
	return nil
}

func newESS() (Asset, error) {
	configPath := "./ess_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Asset{}, err
	}

	machineConfig := MachineConfig{}
	err = json.Unmarshal(jsonConfig, &machineConfig)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	broadcast := make(map[uuid.UUID]chan<- asset.AssetStatus)
	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}
	device := &DummyDevice{}

	return Asset{&sync.Mutex{}, PID, device, broadcast, supervisory, config}, err
}

func TestReadConfig(t *testing.T) {
	testConfig := MachineConfig{}
	jsonConfig, err := ioutil.ReadFile("./ess_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	assertConfig := MachineConfig{"TEST_Virtual ESS", "Virtual Bus", 20, 10, 50}
	assert.Assert(t, testConfig == assertConfig)
}

func TestWriteControl(t *testing.T) {
	ess, err := newESS()
	if err != nil {
		log.Fatal(err)
	}

	control := MachineControl{false, 10, 10, false}
	ess.WriteControl(control)
}

type subscriber struct {
	pid uuid.UUID
	ch  <-chan asset.AssetStatus
}

func TestUpdateStatus(t *testing.T) {
	ess, err := newESS()
	if err != nil {
		log.Fatal(err)
	}

	pid, _ := uuid.NewUUID()
	ch := ess.Subscribe(pid)
	sub := subscriber{pid, ch}

	ess.UpdateStatus()

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	status, ok := <-sub.ch

	assert.Assert(t, ok == true)
	assert.Assert(t, status == assertedStatus)
}

func TestBroadcast(t *testing.T) {
	ess, err := newESS()
	if err != nil {
		log.Fatal(err)
	}

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch := ess.Subscribe(pid)
		subs[i] = subscriber{pid, ch}

	}

	go ess.UpdateStatus()

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	for _, sub := range subs {
		status, ok := <-sub.ch
		assert.Assert(t, ok == true)
		assert.Assert(t, status == assertedStatus)
	}
}

func TestTransform(t *testing.T) {

}
