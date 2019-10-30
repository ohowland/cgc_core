package virtualess

import (
	"testing"
	"time"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
	"gotest.tools/assert"
)

func newESS() ess.Asset {
	configPath := "../ess_test_config.json"
	ess, err := New(configPath)
	if err != nil {
		panic(err)
	}
	return ess
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

	device := ess.DeviceController().(VirtualESS)
	device.LinkToBus(bus)
	return &ess
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

	device := ess.DeviceController().(VirtualESS)
	device.LinkToBus(bus)

	testObservers := bus.GetBusObservers()

	assert.Assert(t, device.observers.AssetObserver == testObservers.AssetObserver)
	assert.Assert(t, device.observers.BusObserver == testObservers.BusObserver)
}

func TestStartVirtualLoop(t *testing.T) {
	ess := newLinkedESS()
	device := ess.DeviceController().(VirtualESS)

	device.StartVirualDevice()
	time.Sleep(1 * time.Second)
	device.StopVirtualDevice()
}

/*
func TestTransitionOffToPQ(t *testing.T) {}

func TestTransitionPQToOff(t *testing.T) {}

func TestTransitionOffToHzV(t *testing.T) {}

func TestTransitionHzVToOff(t *testing.T) {}

func TestTransitionPQToHzV(t *testing.T) {}

func TestTransitionHzVToPQ(t *testing.T) {}

func TestStopVirtualLoop(t *testing.T) {}
*/
