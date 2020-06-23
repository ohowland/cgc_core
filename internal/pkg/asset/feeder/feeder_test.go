package feeder

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"gotest.tools/assert"
)

type DummyDevice struct {
	CloseFeeder bool // control
}

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

func (d *DummyDevice) WriteDeviceControl(ctrl MachineControl) error {
	d.CloseFeeder = ctrl.CloseFeeder
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (d *DummyDevice) Stop() error {
	return nil
}

func newFeeder() (Asset, error) {
	configPath := "./feeder_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Asset{}, err
	}

	return New(jsonConfig, &DummyDevice{})
}

func TestReadConfigFile(t *testing.T) {
	testConfig := StaticConfig{}
	jsonConfig, err := ioutil.ReadFile("./feeder_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	assert.NilError(t, err)

	assertConfig := StaticConfig{"TEST_Virtual Feeder", "Virtual Bus", 20, 18}
	assert.Assert(t, testConfig == assertConfig)
}

func TestReadConfigMem(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	assert.Equal(t, feeder.PID(), feeder.pid)
	assert.Equal(t, feeder.Name(), "TEST_Virtual Feeder")
	assert.Equal(t, feeder.BusName(), "Virtual Bus")
}

func TestRequestControl(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)

	err = feeder.RequestControl(pid, write)
	assert.NilError(t, err)
}

func TestWriteControl(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)
	_ = feeder.RequestControl(pid, write)

	control := MachineControl{true}
	write <- msg.New(pid, msg.Control, control)

	device := feeder.DeviceController().(*DummyDevice)

	if device.CloseFeeder != control.CloseFeeder {
		t.Errorf("TestWriteControl() pass1: FAILED, %v != %v", device.CloseFeeder, control.CloseFeeder)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %v == %v", device.CloseFeeder, control.CloseFeeder)
	}

	control = MachineControl{false}
	write <- msg.New(pid, msg.Control, control)
	if device.CloseFeeder != control.CloseFeeder {
		t.Errorf("TestWriteControl() pass1: FAILED, %v != %v", device.CloseFeeder, control.CloseFeeder)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %v == %v", device.CloseFeeder, control.CloseFeeder)
	}

	close(write)
}

type subscriber struct {
	pid uuid.UUID
	ch  <-chan msg.Msg
}

func TestUpdateStatus(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	ch, err := feeder.Subscribe(pid, msg.Status)
	assert.NilError(t, err)
	sub := subscriber{pid, ch}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg, ok := <-sub.ch
		status := msg.Payload().(Status)

		assertedStatus := Status{
			CalculatedStatus{},
			assertedStatus(),
		}

		assert.Assert(t, ok == true)
		assert.Assert(t, status == assertedStatus)
	}()

	feeder.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherStatus(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := feeder.Subscribe(pid, msg.Status)
		assert.NilError(t, err)
		subs[i] = subscriber{pid, ch}

	}

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	var wg sync.WaitGroup
	for _, sub := range subs {
		wg.Add(1)
		go func(sub subscriber, wg *sync.WaitGroup) {
			defer wg.Done()
			msg, ok := <-sub.ch
			status := msg.Payload().(Status)
			assert.Assert(t, ok == true)
			assert.Assert(t, status == assertedStatus)
		}(sub, &wg)
	}

	feeder.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherConfig(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := feeder.Subscribe(pid, msg.Config)
		assert.NilError(t, err)
		subs[i] = subscriber{pid, ch}
	}

	assertConfig := StaticConfig{"TEST_Virtual Feeder", "Virtual Bus", 20, 18}

	var wg sync.WaitGroup
	for _, sub := range subs {
		wg.Add(1)
		go func(sub subscriber, wg *sync.WaitGroup) {
			defer wg.Done()
			msg, ok := <-sub.ch
			config := msg.Payload().(StaticConfig)
			assert.Assert(t, ok == true)
			assert.Equal(t, config, assertConfig)
		}(sub, &wg)
	}

	feeder.UpdateConfig()
	wg.Wait()
}

func TestUnsubscribeFromPublisher(t *testing.T) {
	feeder, err := newFeeder()
	assert.NilError(t, err)

	rand.Seed(time.Now().UnixNano())
	nSubs := rand.Intn(9) + 1
	subs := make([]subscriber, nSubs)
	for i := 0; i < nSubs; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := feeder.Subscribe(pid, msg.Status)
		assert.NilError(t, err)
		subs[i] = subscriber{pid, ch}
	}

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	unsub := rand.Intn(nSubs)
	feeder.Unsubscribe(subs[unsub].pid)

	var wg sync.WaitGroup
	for i, sub := range subs {
		wg.Add(1)
		go func(sub subscriber, i int, wg *sync.WaitGroup) {
			defer wg.Done()
			msg, ok := <-sub.ch
			if !ok {
				assert.Assert(t, i == unsub)
				return
			}
			status := msg.Payload().(Status)
			assert.Assert(t, status == assertedStatus)
		}(sub, i, &wg)
	}

	feeder.UpdateStatus()
	wg.Wait()
}

func TestTransform(t *testing.T) {
	machineStatus := assertedStatus()

	assertedStatus := Status{
		CalculatedStatus{},
		machineStatus,
	}

	status := transform(machineStatus)
	assert.Assert(t, status == assertedStatus)
}

func TestStatusAccessors(t *testing.T) {
	machineStatus := assertedStatus()

	assertedStatus := Status{
		CalculatedStatus{},
		machineStatus,
	}

	assert.Equal(t, assertedStatus.KW(), machineStatus.KW)
	assert.Equal(t, assertedStatus.KVAR(), machineStatus.KVAR)
}
