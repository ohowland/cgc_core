package virtualfeeder

import (
	"testing"
	"time"

	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
	"gotest.tools/assert"
)

func newFeeder() *feeder.Asset {
	configPath := "../feeder_test_config.json"
	feeder, err := New(configPath)
	if err != nil {
		panic(err)
	}
	return &feeder
}

func newBus() virtualacbus.VirtualACBus {
	configPath := "../../../bus/virtualacbus/virtualacbus_test_config.json"
	bus, err := virtualacbus.New(configPath)
	if err != nil {
		panic(err)
	}
	return bus
}

func newLinkedFeeder() *feeder.Asset {
	feeder := newFeeder()
	bus := newBus()

	device := feeder.DeviceController().(*VirtualFeeder)
	device.LinkToBus(bus)
	return feeder
}

func TestNew(t *testing.T) {
	configPath := "../feeder_test_config.json"
	feeder, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, feeder.Name() == "TEST_Virtual Feeder")
}

func TestLinkToBus(t *testing.T) {
	feeder := newFeeder()
	bus := newBus()

	device := feeder.DeviceController().(*VirtualFeeder)
	device.LinkToBus(bus)

	testObservers := bus.GetBusObservers()

	assert.Assert(t, device.observers.AssetObserver != nil)
	assert.Assert(t, device.observers.BusObserver != nil)

	assert.Assert(t, device.observers.AssetObserver == testObservers.AssetObserver)
	assert.Assert(t, device.observers.BusObserver == testObservers.BusObserver)
}

func TestStartStopVirtualLoop(t *testing.T) {
	feeder := newLinkedFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)

	device.StartVirtualDevice()
	time.Sleep(100 * time.Millisecond)
	device.StopVirtualDevice()
	time.Sleep(100 * time.Millisecond)

	//TODO: What are the conditions for success and failure?
}

func TestRead(t *testing.T) {
	feeder := newLinkedFeeder()
	device := feeder.DeviceController().(*VirtualFeeder)

	device.StartVirtualDevice()
	time.Sleep(1 * time.Second)
	status := device.read()
	time.Sleep(1 * time.Second)
	timestamp := status.timestamp

	assertedStatus := Status{
		timestamp: timestamp,
		KW:        0,
		KVAR:      0,
		Hz:        0,
		Volt:      0,
		Online:    false,
	}

	assert.Assert(t, assertedStatus == status)

	device.StopVirtualDevice()
}
