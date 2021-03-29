package virtualbms

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ohowland/cgc_core/internal/lib/bus/dc/virtualdcbus"
	"github.com/ohowland/cgc_core/internal/pkg/asset/bms"
	"github.com/ohowland/cgc_core/internal/pkg/bus/dc"
	"gotest.tools/assert"
)

// randStatus returns a closure for random DummyAsset Status
func randStatus() func() Status {
	status := Status{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}
	return func() Status {
		return status
	}
}

var assertedStatus = randStatus()

func newBMS() bms.Asset {
	configPath := "../../../../pkg/asset/bms/bms_test_config.json"
	bms, err := New(configPath)
	if err != nil {
		panic(err)
	}
	return bms
}

func newBus() dc.Bus {
	configPath := "../../../../pkg/bus/dc/dc_test_config.json"
	bus, err := virtualdcbus.New(configPath)
	if err != nil {
		panic(err)
	}
	return bus
}

func newLinkedBMS(bus *virtualdcbus.VirtualDCBus) *bms.Asset {
	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)
	bus.AddMember(device)
	return &bms
}

func TestNew(t *testing.T) {
	bms := newBMS()
	assert.Assert(t, bms.Name() == "TEST_Virtual BMS")
}

func TestLinkToVirtualBus(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)
	defer device.Stop()

	relay.AddMember(device)

	assert.Assert(t, device.bus.send != nil)
	assert.Assert(t, device.bus.recieve != nil)

	targetSend := Target{
		pid: device.PID(),
		status: Status{
			KW:                   1,
			Volts:                480,
			SOC:                  0.6,
			RealPositiveCapacity: 10,
			RealNegativeCapacity: 10,
			Online:               true,
		},
	}

	device.bus.send <- targetSend
	time.Sleep(100 * time.Millisecond)
	targetRecieve := <-device.bus.recieve

	// Gridforming device kw/kvar values are not counted in the aggregation.
	// This test will fail if the asset is gridforming.

	// BMS is always grid-forming.
	assert.Assert(t, targetRecieve.KW() == 0)
}

func TestStartStop(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)

	relay.AddMember(device)
	device.Stop()

	_, ok := <-device.comm.send
	assert.Assert(t, !ok)

}

func TestRead(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)
	defer device.Stop()

	relay.AddMember(device)

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}

	assertedStatus := Status{
		KW:                   0,
		Volts:                0,
		SOC:                  0,
		RealPositiveCapacity: 0,
		RealNegativeCapacity: 0,
		Online:               false,
	}

	assert.Assert(t, assertedStatus == status)
}

func TestWrite(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)
	defer device.Stop()

	relay.AddMember(device)

	intercept := make(chan Control)
	device.comm.send = intercept

	control := Control{true, 10}

	go func() {
		err := device.write(control)
		if err != nil {
			t.Error(err)
		}
	}()

	testControl := <-intercept
	assert.Assert(t, testControl == control)
}

func TestReadDeviceStatus(t *testing.T) {
	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	newBMS := newBMS()
	device := newBMS.DeviceController().(*VirtualBMS)
	defer device.Stop()

	relay.AddMember(device)

	machineStatus, _ := device.ReadDeviceStatus()

	assertedStatus := bms.MachineStatus{
		KW:                   0,
		Volts:                0,
		SOC:                  0,
		RealPositiveCapacity: 0,
		RealNegativeCapacity: 0,
		Online:               false,
	}

	assert.Assert(t, machineStatus == assertedStatus)
}

func TestWriteDeviceControl(t *testing.T) {
	newBMS := newBMS()
	device := newBMS.DeviceController().(*VirtualBMS)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	relay.AddMember(device)

	intercept := make(chan Control)
	device.comm.send = intercept

	machineControl := bms.MachineControl{
		Run: true,
		KW:  10,
	}

	go func() {
		err := device.WriteDeviceControl(machineControl)
		if err != nil {
			t.Error(err)
		}
	}()

	testControl := <-intercept
	assert.Assert(t, testControl == mapControl(machineControl))
}

func TestMapStatus(t *testing.T) {
	status := Status{
		KW:                   1,
		Volts:                4,
		SOC:                  5,
		RealPositiveCapacity: 6,
		RealNegativeCapacity: 7,
		Online:               true,
	}

	machineStatus := mapStatus(status)

	assertedStatus := bms.MachineStatus{
		KW:                   1,
		Volts:                4,
		SOC:                  5,
		RealPositiveCapacity: 6,
		RealNegativeCapacity: 7,
		Online:               true,
	}

	assert.Assert(t, machineStatus == assertedStatus)
}

func TestMapControl(t *testing.T) {

	machineControl := bms.MachineControl{
		Run: true,
		KW:  10,
	}

	Control := Control{
		Run: true,
		KW:  10,
	}

	assert.Assert(t, mapControl(machineControl) == Control)
}

func TestTransitionOffToOn(t *testing.T) {
	/*
		This test currently fails. Inverter cannot enter PQ unlbms the bus is energized.
		Need mock the virtual bus as always energized.
	*/
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	relay.AddMember(device)

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}

	assert.Assert(t, status.Online == false)

	control := Control{
		Run: true,
		KW:  0,
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

	bms := newBMS()
	device := bms.DeviceController().(*VirtualBMS)
	defer device.Stop()

	bus := newBus()
	relay := bus.Relayer().(*virtualdcbus.VirtualDCBus)

	relay.AddMember(device)

	status, err := device.read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Assert(t, status.Online == false)

	control := Control{
		Run: true,
		KW:  0,
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
		Run: false,
		KW:  0,
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
