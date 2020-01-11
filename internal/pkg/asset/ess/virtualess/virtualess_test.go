package virtualess

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus/virtualacbus"
	"gotest.tools/assert"
)

// randStatus returns a closure for random DummyAsset Status
func randStatus() func() Status {
	status := Status{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false, false}
	return func() Status {
		return status
	}
}

var assertedStatus = randStatus()

func newESS() ess.Asset {
	configPath := "../ess_test_config.json"
	ess, err := New(configPath)
	if err != nil {
		panic(err)
	}
	return ess
}

func newBus() acbus.ACBus {
	configPath := "../../../bus/acbus/acbus_test_config.json"
	bus, err := virtualacbus.New(configPath)
	if err != nil {
		panic(err)
	}
	return bus
}

func newLinkedESS(bus *virtualacbus.VirtualACBus) *ess.Asset {
	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)
	bus.AddMember(device)
	return &ess
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
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	assert.Assert(t, device.bus.send != nil)
	assert.Assert(t, device.bus.recieve != nil)

	targetSend := Target{
		pid: device.PID(),
		status: Status{
			KW:   1,
			KVAR: 2,
			Hz:   60,
			Volt: 480,
		},
	}

	device.bus.send <- targetSend
	targetRecieve := <-device.bus.recieve

	assert.Assert(t, targetSend.KW() == -1*targetRecieve.KW())
	assert.Assert(t, targetSend.KVAR() == targetRecieve.KVAR())

	relay.RemoveMember(device.PID())
}

func TestStartStopProcess(t *testing.T) {
	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)
	device.StopProcess()

	_, ok := <-device.comm.outgoing
	assert.Assert(t, !ok)
}

func TestRead(t *testing.T) {
	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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
		RealPositiveCapacity: 0,
		RealNegativeCapacity: 0,
		Gridforming:          false,
		Online:               false,
	}

	assert.Assert(t, assertedStatus == status)

	relay.RemoveMember(device.PID())
}

func TestWrite(t *testing.T) {
	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

func TestReadDeviceStatus(t *testing.T) {
	newESS := newESS()
	device := newESS.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	machineStatus, _ := device.ReadDeviceStatus()

	assertedStatus := ess.MachineStatus{
		KW:                   0,
		KVAR:                 0,
		Hz:                   0,
		Volt:                 0,
		SOC:                  0,
		RealPositiveCapacity: 0,
		RealNegativeCapacity: 0,
		Gridforming:          false,
		Online:               false,
	}

	assert.Assert(t, machineStatus == assertedStatus)

}

func TestWriteDeviceControl(t *testing.T) {
	newESS := newESS()
	device := newESS.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

	intercept := make(chan Control)
	device.comm.outgoing = intercept

	machineControl := ess.MachineControl{true, 10, 10, false}

	go func() {
		err := device.WriteDeviceControl(machineControl)
		if err != nil {
			t.Fatal(err)
		}
	}()

	testControl := <-intercept
	assert.Assert(t, testControl == mapControl(machineControl))
}

func TestMapStatus(t *testing.T) {
	status := Status{
		KW:                   1,
		KVAR:                 2,
		Hz:                   3,
		Volt:                 4,
		SOC:                  5,
		RealPositiveCapacity: 6,
		RealNegativeCapacity: 7,
		Gridforming:          true,
		Online:               true,
	}

	machineStatus := mapStatus(status)

	assertedStatus := ess.MachineStatus{
		KW:                   1,
		KVAR:                 2,
		Hz:                   3,
		Volt:                 4,
		SOC:                  5,
		RealPositiveCapacity: 6,
		RealNegativeCapacity: 7,
		Gridforming:          true,
		Online:               true,
	}

	assert.Assert(t, machineStatus == assertedStatus)

}

func TestMapControl(t *testing.T) {

	machineControl := ess.MachineControl{
		Run:      true,
		KW:       10,
		KVAR:     11,
		Gridform: false,
	}

	Control := Control{
		Run:      true,
		KW:       10,
		KVAR:     11,
		Gridform: false,
	}

	assert.Assert(t, mapControl(machineControl) == Control)
}

func TestTransitionOffToPQ(t *testing.T) {
	/*
		This test currently fails. Inverter cannot enter PQ unless the bus is energized.
		Need mock the virtual bus as always energized.
	*/
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

	assert.Assert(t, status.Online == false)
	assert.Assert(t, status.Gridforming == false)

	relay.RemoveMember(device.PID())
}

func TestTransitionPQToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

	relay.RemoveMember(device.PID())
}

func TestTransitionOffToHzV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

	relay.RemoveMember(device.PID())
}

func TestTransitionHzVToOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

	relay.RemoveMember(device.PID())
}

func TestTransitionPQToHzV(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

	relay.RemoveMember(device.PID())
}

func TestTransitionHzVToPQ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ess := newESS()
	device := ess.DeviceController().(*VirtualESS)

	bus := newBus()
	relay := bus.Relayer().(*virtualacbus.VirtualACBus)

	relay.AddMember(device)

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

	relay.RemoveMember(device.PID())
}
