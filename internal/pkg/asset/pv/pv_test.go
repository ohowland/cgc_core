package pv

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
	KW  float64 // control
	Run bool    // control
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
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return assertedStatus(), nil
}

func (d *DummyDevice) WriteDeviceControl(ctrl MachineControl) error {
	d.KW = ctrl.KWLimit
	d.Run = ctrl.Run
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return nil
}

func (d *DummyDevice) Stop() error {
	return nil
}

func newPV() (Asset, error) {
	configPath := "./pv_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Asset{}, err
	}

	return New(jsonConfig, &DummyDevice{})
}

func TestReadConfigFile(t *testing.T) {
	testConfig := StaticConfig{}
	jsonConfig, err := ioutil.ReadFile("./pv_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	assert.NilError(t, err)

	assertConfig := StaticConfig{"TEST_Virtual PV", "Virtual Bus", 20, 10}
	assert.Assert(t, testConfig == assertConfig)
}

func TestReadConfigMem(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	assert.Equal(t, ess.PID(), ess.pid)
	assert.Equal(t, ess.Name(), "TEST_Virtual PV")
	assert.Equal(t, ess.BusName(), "Virtual Bus")
}

func TestRequestControl(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)

	err = ess.RequestControl(pid, write)
	assert.NilError(t, err)
}

func TestWriteControl(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)
	_ = ess.RequestControl(pid, write)

	control := MachineControl{true, rand.Float64(), rand.Float64()}
	write <- msg.New(pid, msg.Control, control)

	device := ess.DeviceController().(*DummyDevice)

	if device.KW != control.KWLimit {
		t.Errorf("TestWriteControl() pass1: FAILED, %f != %f", device.KW, control.KWLimit)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %f == %f", device.KW, control.KWLimit)
	}

	rand.Seed(time.Now().UnixNano())
	control = MachineControl{true, rand.Float64(), rand.Float64()}
	write <- msg.New(pid, msg.Control, control)
	if device.KW != control.KWLimit {
		t.Errorf("TestWriteControl() pass2: FAILED, %f != %f", device.KW, control.KWLimit)
	} else {
		t.Logf("TestWriteControl() pass2: PASSED, %f == %f", device.KW, control.KWLimit)
	}

	close(write)
}

type subscriber struct {
	pid uuid.UUID
	ch  <-chan msg.Msg
}

func TestUpdateStatus(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	ch, err := ess.Subscribe(pid, msg.Status)
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

	ess.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherStatus(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := ess.Subscribe(pid, msg.Status)
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

	ess.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherConfig(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := ess.Subscribe(pid, msg.Config)
		assert.NilError(t, err)

		subs[i] = subscriber{pid, ch}
	}

	assertConfig := StaticConfig{"TEST_Virtual PV", "Virtual Bus", 20, 10}

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

	ess.UpdateConfig()
	wg.Wait()
}

func TestUnsubscribeFromPublisher(t *testing.T) {
	ess, err := newPV()
	assert.NilError(t, err)

	rand.Seed(time.Now().UnixNano())
	nSubs := rand.Intn(9) + 1
	subs := make([]subscriber, nSubs)
	for i := 0; i < nSubs; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := ess.Subscribe(pid, msg.Status)
		assert.NilError(t, err)

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
	assert.Equal(t, assertedStatus.RealPositiveCapacity(), 0)
	assert.Equal(t, assertedStatus.RealNegativeCapacity(), 0)
}
