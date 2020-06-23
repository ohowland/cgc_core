package grid

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
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return assertedStatus(), nil
}

func (d *DummyDevice) WriteDeviceControl(c MachineControl) error {
	d.CloseIntertie = c.CloseIntertie
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return nil
}

func (d *DummyDevice) Stop() error {
	return nil
}

func newGrid() (Asset, error) {
	configPath := "./grid_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Asset{}, err
	}

	return New(jsonConfig, &DummyDevice{})
}

func TestReadConfigFile(t *testing.T) {
	testConfig := MachineConfig{}
	jsonConfig, err := ioutil.ReadFile("./grid_test_config.json")
	err = json.Unmarshal(jsonConfig, &testConfig)
	if err != nil {
		t.Fatal(err)
	}

	assertConfig := MachineConfig{"TEST_Virtual Grid", "Virtual Bus", 20, 19}
	assert.Assert(t, testConfig == assertConfig)
}

func TestReadConfigMem(t *testing.T) {
	grid, err := newGrid()
	assert.NilError(t, err)

	assert.Equal(t, grid.PID(), grid.pid)
	assert.Equal(t, grid.Name(), "TEST_Virtual Grid")
	assert.Equal(t, grid.BusName(), "Virtual Bus")
}

func TestRequestControl(t *testing.T) {
	grid, err := newGrid()
	assert.NilError(t, err)
	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)

	err = grid.RequestControl(pid, write)
	assert.NilError(t, err)

}

func TestWriteControl(t *testing.T) {
	grid, err := newGrid()
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	write := make(chan msg.Msg)
	_ = grid.RequestControl(pid, write)

	control := MachineControl{true}
	write <- msg.New(pid, msg.Control, control)

	device := grid.DeviceController().(*DummyDevice)

	if device.CloseIntertie != control.CloseIntertie {
		t.Errorf("TestWriteControl() pass1: FAILED, %v != %v", device.CloseIntertie, control.CloseIntertie)
	} else {
		t.Logf("TestWriteControl() pass1: PASSED, %v == %v", device.CloseIntertie, control.CloseIntertie)
	}

	control = MachineControl{false}
	write <- msg.New(pid, msg.Control, control)
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
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	ch, err := grid.Subscribe(pid, msg.Status)
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

	grid.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherStatus(t *testing.T) {
	grid, err := newGrid()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := grid.Subscribe(pid, msg.Status)
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
			wg.Done()
			msg, ok := <-sub.ch
			status := msg.Payload().(Status)
			assert.Assert(t, ok == true)
			assert.Assert(t, status == assertedStatus)
		}(sub, &wg)
	}

	grid.UpdateStatus()
	wg.Wait()
}

func TestSubscribeToPublisherConfig(t *testing.T) {
	grid, err := newGrid()
	assert.NilError(t, err)

	n := 3
	subs := make([]subscriber, n)
	for i := 0; i < n; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := grid.Subscribe(pid, msg.Config)
		assert.NilError(t, err)
		subs[i] = subscriber{pid, ch}
	}

	assertConfig := MachineConfig{"TEST_Virtual Grid", "Virtual Bus", 20, 19}

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

	grid.UpdateConfig()
	wg.Wait()
}

func TestUnsubscribeFromPublisher(t *testing.T) {
	grid, err := newGrid()
	assert.NilError(t, err)

	rand.Seed(time.Now().UnixNano())
	nSubs := rand.Intn(9) + 1
	subs := make([]subscriber, nSubs)
	for i := 0; i < nSubs; i++ {
		pid, _ := uuid.NewUUID()
		ch, err := grid.Subscribe(pid, msg.Status)
		assert.NilError(t, err)
		subs[i] = subscriber{pid, ch}
	}

	assertedStatus := Status{
		CalculatedStatus{},
		assertedStatus(),
	}

	unsub := rand.Intn(nSubs)
	grid.Unsubscribe(subs[unsub].pid)

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

	grid.UpdateStatus()
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
