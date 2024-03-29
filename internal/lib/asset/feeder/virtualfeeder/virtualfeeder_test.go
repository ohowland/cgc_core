package virtualfeeder

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ohowland/cgc_core/internal/lib/bus/ac/virtualacbus"
	"github.com/ohowland/cgc_core/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc_core/internal/pkg/bus/ac"
	"gotest.tools/assert"
)

// randStatus returns a closure for random DummyAsset Status
func randStatus() func() Status {
	status := Status{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}
	return func() Status {
		return status
	}
}

var assertedStatus = randStatus()

func newFeeder() feeder.Asset {
	configPath := "../../../../pkg/asset/feeder/feeder_test_config.json"
	feeder, err := New(configPath)
	if err != nil {
		panic(err)
	}
	return feeder
}

func newBus() ac.Bus {
	configPath := "../../../../pkg/bus/ac/ac_test_config.json"
	bus, err := virtualacbus.New(configPath)
	if err != nil {
		panic(err)
	}
	return bus
}

func newLinkedFeeder(bus *virtualacbus.VirtualACBus) *feeder.Asset {
	feeder := newFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)
	bus.AddMember(device)
	return &feeder
}

func TestNew(t *testing.T) {
	configPath := "../../../../pkg/asset/feeder/feeder_test_config.json"
	feeder, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, feeder.Name() == "TEST_Virtual Feeder")
}

func TestLinkToVirtualBus(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	feeder := newFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	relay.AddMember(device)

	assert.Assert(t, device.bus.send != nil)
	assert.Assert(t, device.bus.recieve != nil)

	targetSend := Target{
		pid: device.PID(),
		status: Status{
			KW:    1,
			KVAR:  2,
			Hz:    60,
			Volts: 480,
		},
	}

	device.bus.send <- targetSend
	time.Sleep(100 * time.Millisecond)
	targetRecieve := <-device.bus.recieve

	assert.Assert(t, targetSend.KW() == -1*targetRecieve.KW())
	assert.Assert(t, targetSend.KVAR() == targetRecieve.KVAR())
}

func TestStartStopProcess(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	feeder := newFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)

	relay.AddMember(device)
	device.Stop()

	_, ok := <-device.comm.send
	assert.Assert(t, !ok)
}

func TestRead(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	feeder := newFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	relay.AddMember(device)

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}

	assertedStatus := Status{
		KW:     0,
		KVAR:   0,
		Hz:     0,
		Volts:  0,
		Online: false,
	}

	assert.Assert(t, assertedStatus == status)
}

func TestWrite(t *testing.T) {
	feeder := newFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	intercept := make(chan Control)
	device.comm.send = intercept

	control := Control{true}

	go func() {
		err := device.write(control)
		assert.NilError(t, err)
	}()

	testControl := <-intercept
	assert.Assert(t, testControl == control)
}

func TestReadDeviceStatus(t *testing.T) {
	newfeeder := newFeeder()
	device := newfeeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	machineStatus, _ := device.ReadDeviceStatus()

	assertedStatus := feeder.MachineStatus{
		KW:     0,
		KVAR:   0,
		Hz:     0,
		Volts:  0,
		Online: false,
	}

	assert.Assert(t, machineStatus == assertedStatus)
}

func TestWriteDeviceControl(t *testing.T) {
	newfeeder := newFeeder()
	device := newfeeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	intercept := make(chan Control)
	device.comm.send = intercept

	machineControl := feeder.MachineControl{CloseFeeder: true}

	go func() {
		err := device.WriteDeviceControl(machineControl)
		assert.NilError(t, err)
	}()

	testControl := <-intercept
	assert.Assert(t, testControl == mapControl(machineControl))
}

func TestMapStatus(t *testing.T) {
	status := Status{
		KW:     1,
		KVAR:   2,
		Hz:     3,
		Volts:  4,
		Online: true,
	}

	machineStatus := mapStatus(status)

	assertedStatus := feeder.MachineStatus{
		KW:     1,
		KVAR:   2,
		Hz:     3,
		Volts:  4,
		Online: true,
	}

	assert.Assert(t, machineStatus == assertedStatus)
}

func TestMapControl(t *testing.T) {

	machineControl := feeder.MachineControl{
		CloseFeeder: true,
	}

	Control := Control{
		CloseFeeder: true,
	}

	assert.Assert(t, mapControl(machineControl) == Control)
}

func TestTransitionOffToOn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	newFeeder := newFeeder()
	device := newFeeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, status.Online == false)

	control := Control{
		CloseFeeder: true,
	}

	err = device.write(control)
	if err != nil {
		t.Fatal(err)
	}

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
}

func TestTransitionOnToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	newFeeder := newFeeder()
	device := newFeeder.DeviceController().(*VirtualFeeder)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)

	control := Control{
		CloseFeeder: true,
	}

	err = device.write(control)
	if err != nil {
		t.Fatal(err)
	}

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)

	control = Control{
		CloseFeeder: false,
	}

	err = device.write(control)
	if err != nil {
		t.Fatal(err)
	}

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == false)
}
