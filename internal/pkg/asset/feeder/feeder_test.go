package feeder

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
	status := MachineStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}
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

func newFeeder() (Asset, error) {
	configPath := "./feeder_test_config.json"
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

	broadcast := make(map[uuid.UUID]chan<- asset.Status)
	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}
	device := &DummyDevice{}

	return Asset{&sync.Mutex{}, PID, device, broadcast, supervisory, config}, err
}

func TestReadConfig(t *testing.T) {
	testConfig := MachineConfig{}
	jsonConfig, err := ioutil.ReadFile("./feeder_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	assertConfig := MachineConfig{"TEST_Virtual Feeder", "Virtual Bus", 20, 18}
	assert.Assert(t, testConfig == assertConfig)
}

func TestWriteControl(t *testing.T) {
	feeder, err := newFeeder()
	if err != nil {
		log.Fatal(err)
	}

	control := MachineControl{false}
	feeder.WriteControl(control)
}

type subscriber struct {
	pid uuid.UUID
	ch  <-chan asset.Status
}

func TestUpdateStatus(t *testing.T) {
	feeder, err := newFeeder()
	if err != nil {
		log.Fatal(err)
	}

	pid, _ := uuid.NewUUID()
	ch := feeder.Subscribe(pid)
	sub := subscriber{pid, ch}

	feeder.UpdateStatus()

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	status, ok := <-sub.ch

	assert.Assert(t, ok == true)
	assert.Assert(t, status == assertedStatus)
}

func TestBroadcast(t *testing.T) {
	feeder, err := newFeeder()
	if err != nil {
		log.Fatal(err)
	}

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch := feeder.Subscribe(pid)
		subs[i] = subscriber{pid, ch}

	}

	go feeder.UpdateStatus()

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
