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

	assert.Assert(t, ess.Name() == "TEST_Virtual ESS")
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

	device.StartVirualDevice()
	time.Sleep(100 * time.Millisecond)
	device.StopVirtualDevice()
	time.Sleep(100 * time.Millisecond)

	// TODO: What are the conditions for success and failure?
}

func TestTransitionOffToPQ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirualDevice()

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl := ess.DispatchControlHandle()
	ctrl.RunCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == false)

	device.StopVirtualDevice()
}

func TestTransitionPQToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirualDevice()

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl := ess.DispatchControlHandle()
	ctrl.RunCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl.RunCmd(false)
	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	device.StopVirtualDevice()
}

func TestTransitionOffToHzV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirualDevice()

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl := ess.DispatchControlHandle()
	ctrl.RunCmd(true)
	ctrl.GridformCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == true)

	device.StopVirtualDevice()
}

func TestTransitionHzVToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirualDevice()

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl := ess.DispatchControlHandle()
	ctrl.RunCmd(true)
	ctrl.GridformCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == true)

	ctrl.RunCmd(false)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	device.StopVirtualDevice()
}

func TestTransitionPQToHzV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirualDevice()

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl := ess.DispatchControlHandle()
	ctrl.RunCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl.GridformCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == true)

	device.StopVirtualDevice()
}

func TestTransitionHzVToPQ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newLinkedESS()
	device := ess.DeviceController().(*VirtualESS)
	device.StartVirualDevice()

	assert.Assert(t, device.status.Online == false)
	assert.Assert(t, device.status.Gridforming == false)

	ctrl := ess.DispatchControlHandle()
	ctrl.RunCmd(true)
	ctrl.GridformCmd(true)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == true)

	ctrl.GridformCmd(false)

	ess.WriteControl()
	ess.UpdateStatus() // virtual device fuzzing requires multiple reads
	ess.UpdateStatus()
	ess.UpdateStatus()
	time.Sleep(6 * time.Second)

	assert.Assert(t, device.status.Online == true)
	assert.Assert(t, device.status.Gridforming == false)

	device.StopVirtualDevice()
}
