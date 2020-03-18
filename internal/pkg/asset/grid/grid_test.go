package grid

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
	"gotest.tools/assert"
)

type DummyDevice struct {
	CloseIntertie bool
}

// randMachineStatus returns a closure for random MachineStatus
func randMachineStatus() func() MachineStatus {
	status := MachineStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}
	return func() MachineStatus {
		return status
	}
}

var assertedStatus = randMachineStatus()

func (d DummyDevice) ReadDeviceStatus() (MachineStatus, error) {
	time.Sleep(100 * time.Millisecond) // fuzz
	return assertedStatus(), nil
}

func (d *DummyDevice) WriteDeviceControl(c MachineControl) error {
	d.CloseIntertie = c.CloseIntertie
	time.Sleep(100 * time.Millisecond) // fuzz
	return nil
}

func newGrid() (Asset, error) {
	configPath := "./grid_test_config.json"
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

	broadcast := make(map[uuid.UUID]chan<- msg.Msg)

	controlOwner := PID

	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}
	device := &DummyDevice{}

	return Asset{&sync.Mutex{}, PID, device, broadcast, controlOwner, supervisory, config}, err
}

func TestReadConfig(t *testing.T) {
	testConfig := MachineConfig{}
	jsonConfig, err := ioutil.ReadFile("./grid_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	assertConfig := MachineConfig{"TEST_Virtual Grid", "Virtual Bus", 20, 19}
	assert.Assert(t, testConfig == assertConfig)
}

func TestRequestControl(t *testing.T) {
	grid1, err := newGrid()
	if err != nil {
		log.Fatal(err)
	}

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)

	ok := grid1.RequestControl(pid, write)
	if !ok {
		t.Error("RequestControl(): FAILED, RequestControl() returned false")
	} else {
		t.Log("RequestControl(): PASSED, RequestControl returned true")
	}
}

func TestWriteControl(t *testing.T) {
	grid1, err := newGrid()
	if err != nil {
		log.Fatal(err)
	}

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)
	_ = grid1.RequestControl(pid, write)

	control := MachineControl{true}
	write <- msg.New(pid, msg.CONTROL, control)

	device := grid1.DeviceController().(*DummyDevice)

	if device.CloseIntertie != control.CloseIntertie {
		t.Errorf("TestWriteControl() pass1: FAILED, %v != %v", device.CloseIntertie, control.CloseIntertie)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %v == %v", device.CloseIntertie, control.CloseIntertie)
	}

	rand.Seed(42)
	control = MachineControl{false}
	write <- msg.New(pid, msg.CONTROL, control)
	if device.CloseIntertie != control.CloseIntertie {
		t.Errorf("TestWriteControl() pass1: FAILED, %v != %v", device.CloseIntertie, control.CloseIntertie)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %v == %v", device.CloseIntertie, control.CloseIntertie)
	}

	close(write)
}

type subscriber struct {
	pid uuid.UUID
	ch  <-chan msg.Msg
}

func TestUpdateStatus(t *testing.T) {
	grid, err := newGrid()
	if err != nil {
		log.Fatal(err)
	}

	pid, _ := uuid.NewUUID()
	ch := grid.Subscribe(pid)
	sub := subscriber{pid, ch}

	go func() {
		msg, ok := <-sub.ch
		status := msg.Payload().(Status)

		assertedStatus := Status{
			CalculatedStatus{},
			assertedStatus(),
		}

		assert.Assert(t, ok == true)
		assert.Assert(t, status == assertedStatus)
	}()

	grid.UpdateStatus()
}

func TestBroadcast(t *testing.T) {
	grid, err := newGrid()
	if err != nil {
		log.Fatal(err)
	}

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch := grid.Subscribe(pid)
		subs[i] = subscriber{pid, ch}

	}

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	for _, sub := range subs {
		go func(sub subscriber) {
			msg, ok := <-sub.ch
			status := msg.Payload().(Status)
			assert.Assert(t, ok == true)
			assert.Assert(t, status == assertedStatus)
		}(sub)
	}

	grid.UpdateStatus()
}

func TestTransform(t *testing.T) {

}
