package ess

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
	"gotest.tools/assert"
)

type DummyDevice struct {
	KW  float64 // control
	Run bool    // control
}

// randMachineStatus returns a closure for random MachineStatus
func randMachineStatus() func() MachineStatus {
	status := MachineStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false, false}
	return func() MachineStatus {
		return status
	}
}

var assertedStatus = randMachineStatus()

func (d DummyDevice) ReadDeviceStatus() (MachineStatus, error) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return assertedStatus(), nil
}

func (d *DummyDevice) WriteDeviceControl(ctrl MachineControl) error {
	d.KW = ctrl.KW
	d.Run = ctrl.Run
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return nil
}

func newESS() (Asset, error) {
	configPath := "./ess_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Asset{}, err
	}

	return New(jsonConfig, &DummyDevice{})
}

func TestReadConfigFile(t *testing.T) {
	testConfig := MachineConfig{}
	jsonConfig, err := ioutil.ReadFile("./ess_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	assert.NilError(t, err)

	assertConfig := MachineConfig{"TEST_Virtual ESS", "Virtual Bus", 20, 10, 50}
	assert.Assert(t, testConfig == assertConfig)
}

func TestReadConfigMem(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	assert.Equal(t, ess.PID(), ess.pid)
	assert.Equal(t, ess.Name(), "TEST_Virtual ESS")
	assert.Equal(t, ess.BusName(), "Virtual Bus")
}

func TestRequestControl(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)

	ok := ess.RequestControl(pid, write)
	assert.Equal(t, ok, true, "RequestControl failed to return ok==true")
}

func TestWriteControl(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)
	_ = ess.RequestControl(pid, write)

	control := MachineControl{true, rand.Float64(), rand.Float64(), true}
	write <- msg.New(pid, control)

	device := ess.DeviceController().(*DummyDevice)

	if device.KW != control.KW {
		t.Errorf("TestWriteControl() pass1: FAILED, %f != %f", device.KW, control.KW)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %f == %f", device.KW, control.KW)
	}

	rand.Seed(time.Now().UnixNano())
	control = MachineControl{true, rand.Float64(), rand.Float64(), true}
	write <- msg.New(pid, control)
	if device.KW != control.KW {
		t.Errorf("TestWriteControl() pass2: FAILED, %f != %f", device.KW, control.KW)
	} else {
		t.Logf("TestWriteControl() pass2: PASSED, %f == %f", device.KW, control.KW)
	}

	close(write)
}

type subscriber struct {
	pid uuid.UUID
	ch  <-chan msg.Msg
}

func TestUpdateStatus(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	ch := ess.Subscribe(pid, msg.Status)
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

	ess.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherStatus(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch := ess.Subscribe(pid, msg.Status)
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

	ess.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherConfig(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch := ess.Subscribe(pid, msg.Config)
		subs[i] = subscriber{pid, ch}
	}

	assertConfig := MachineConfig{"TEST_Virtual ESS", "Virtual Bus", 20, 10, 50}

	var wg sync.WaitGroup
	for _, sub := range subs {
		wg.Add(1)
		go func(sub subscriber, wg *sync.WaitGroup) {
			defer wg.Done()
			msg, ok := <-sub.ch
			config := msg.Payload().(MachineConfig)
			assert.Assert(t, ok == true)
			assert.Equal(t, config, assertConfig)
		}(sub, &wg)
	}

	ess.UpdateConfig()
	wg.Wait()
}

func TestUnsubscribeFromPublisher(t *testing.T) {
	ess, err := newESS()
	assert.NilError(t, err)

	rand.Seed(time.Now().UnixNano())
	nSubs := rand.Intn(9) + 1
	subs := make([]subscriber, nSubs)
	for i := 0; i < nSubs; i++ {
		pid, _ := uuid.NewUUID()
		ch := ess.Subscribe(pid, msg.Status)
		subs[i] = subscriber{pid, ch}
	}

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	unsub := rand.Intn(nSubs)
	ess.Unsubscribe(subs[unsub].pid)

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

	ess.UpdateStatus()
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
	assert.Equal(t, assertedStatus.RealPositiveCapacity(), machineStatus.RealPositiveCapacity)
	assert.Equal(t, assertedStatus.RealNegativeCapacity(), machineStatus.RealNegativeCapacity)
}
