package virtualess

import (
	"testing"
	"time"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
	"gotest.tools/assert"
)

func newESS() *ess.Asset {
	configPath := "../ess_test_config.json"
	ess, err := New(configPath)
	if err != nil {
		panic(err)
	}
	return &ess
}

func newBus() virtualacbus.VirtualACBus {
	configPath := "../../../bus/virtualacbus/virtualacbus_test_config.json"
	bus, err := virtualacbus.New(configPath)
	if err != nil {
		panic(err)
	}
	return bus
}

func newLinkedESS() *ess.Asset {
	ess := newESS()
	bus := newBus()

	device := ess.DeviceController().(*VirtualESS)
	device.LinkToBus(bus)
	return ess
}

func TestNew(t *testing.T) {
	configPath := "../ess_test_config.json"
	ess, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, ess.Config().Name() == "TEST_Virtual ESS")
}

func TestLinkToBus(t *testing.T) {

	ess := newESS()
	bus := newBus()

	device := ess.DeviceController().(*VirtualESS)
	device.LinkToBus(bus)

	testObservers := bus.GetBusObservers()

	assert.Assert(t, device.observers.AssetObserver != nil)
	assert.Assert(t, device.observers.BusObserver != nil)

	assert.Assert(t, device.observers.AssetObserver == testObservers.AssetObserver)
	assert.Assert(t, device.observers.BusObserver == testObservers.BusObserver)
}

func TestStartStopVirtualLoop(t *testing.T) {
	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)

	device.StartVirtualDevice()
	time.Sleep(100 * time.Millisecond)
	device.StopVirtualDevice()
	time.Sleep(100 * time.Millisecond)

	// TODO: What are the conditions for success and failure?
}

func TestRead(t *testing.T) {
	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)

	device.StartVirtualDevice()
	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}

	assertedStatus := Status{
		KW:                   0,
		KVAR:                 0,
		Hz:                   0,
		Volt:                 0,
		SOC:                  0,
		PositiveRealCapacity: 0,
		NegativeRealCapacity: 0,
		Gridforming:          false,
		Online:               false,
	}

	assert.Assert(t, assertedStatus == status)

	device.StopVirtualDevice()
}

func TestWrite(t *testing.T) {
	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	intercept := make(chan Control)
	device.comm.outgoing = intercept

	control := Control{true, 10, 10, false}

	go func() {
		err := device.write(control)
		if err != nil {
			t.Fatal(err)
		}
	}()

	testControl := <-intercept
	assert.Assert(t, testControl == control)
}

func TestReadDeviceStatus(t *testing.T) {}

func TestWriteDeviceControl(t *testing.T) {}

func TestUpdateObservers(t *testing.T) {}

func TestMapStatus(t *testing.T) {}

func TestMapControl(t *testing.T) {}

func TestMapSource(t *testing.T) {}

func TestTransitionOffToPQ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirtualDevice()

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	control := Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: false,
	}

	err = device.write(control)
	if err != nil {
		t.Fatal(err)
	}

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == false)

	device.StopVirtualDevice()
}

func TestTransitionPQToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirtualDevice()

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	control := Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: false,
	}

	err = device.write(control)
	if err != nil {
		t.Fatal(err)
	}

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == false)

	control = Control{
		Run:      false,
		KW:       0,
		KVAR:     0,
		Gridform: false,
	}

	err = device.write(control)
	if err != nil {
		t.Fatal(err)
	}

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	device.StopVirtualDevice()
}

func TestTransitionOffToHzV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirtualDevice()

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	control := Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: true,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == true)

	device.StopVirtualDevice()
}

func TestTransitionHzVToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirtualDevice()

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	control := Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: true,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == true)

	control = Control{
		Run:      false,
		KW:       0,
		KVAR:     0,
		Gridform: true,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	device.StopVirtualDevice()
}

func TestTransitionPQToHzV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirtualDevice()

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	control := Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: false,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == false)

	control = Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: true,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == true)

	device.StopVirtualDevice()
}

func TestTransitionHzVToPQ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirtualDevice()

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	control := Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: true,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == true)

	control = Control{
		Run:      true,
		KW:       0,
		KVAR:     0,
		Gridform: false,
	}
	err = device.write(control)

	status, err = device.read()
	time.Sleep(2 * time.Second)
	status, err = device.read()

	assert.Assert(t, status.Online == true)
	assert.Assert(t, status.Gridforming == false)

	device.StopVirtualDevice()
}
