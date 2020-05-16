package virtualfeeder

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
)

// VirtualFeeder target
type VirtualFeeder struct {
	pid  uuid.UUID
	comm virtualHardware
	bus  virtualBus
}

// Comm data structure for the VirtualFeeder
type virtualHardware struct {
	send    chan Control
	recieve chan Status
}

type virtualBus struct {
	send    chan<- asset.VirtualStatus
	recieve <-chan asset.VirtualStatus
}

// Target is a virtual representation of the hardware
type Target struct {
	pid     uuid.UUID
	status  Status
	control Control
}

// KW is an accessor for real power
func (t Target) KW() float64 {
	return t.status.KW
}

// KVAR is an accessor for reactive power
func (t Target) KVAR() float64 {
	return t.status.KVAR
}

// Hz is an accessor for frequency
func (t Target) Hz() float64 {
	return t.status.Hz
}

// Volt is an accessor for ac voltage
func (t Target) Volt() float64 {
	return t.status.Volt
}

// Gridforming is an accessor for gridforming state
func (t Target) Gridforming() bool {
	return false
}

// Status data structure for the VirtualFeeder
type Status struct {
	KW     float64 `json:"KW"`
	KVAR   float64 `json:"KVAR"`
	Hz     float64 `json:"Hz"`
	Volt   float64 `json:"Volts"`
	Online bool    `json:"Online"`
}

// Control data structure for the VirtualFeeder
type Control struct {
	CloseFeeder bool
}

// PID is an accessor for the process id
func (a VirtualFeeder) PID() uuid.UUID {
	return a.pid
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualFeeder) ReadDeviceStatus() (feeder.MachineStatus, error) {
	status, err := a.read()
	return mapStatus(status), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualFeeder) WriteDeviceControl(machineControl feeder.MachineControl) error {
	control := mapControl(machineControl)
	err := a.write(control)
	return err
}

func (a VirtualFeeder) read() (Status, error) {
	fuzzing := rand.Intn(500)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.recieve
	if !ok {
		return Status{}, errors.New("Read Error")
	}
	return readStatus, nil
}

func (a VirtualFeeder) write(control Control) error {
	a.comm.send <- control
	return nil
}

// New returns an initalized VirtualFeeder Asset; this is part of the Asset interface.
func New(configPath string) (feeder.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return feeder.Asset{}, err
	}

	pid, err := uuid.NewUUID()

	device := VirtualFeeder{
		pid:  pid,
		comm: virtualHardware{},
	}

	return feeder.New(jsonConfig, &device)
}

// Status maps feeder.DeviceStatus to feeder.Status
func mapStatus(s Status) feeder.MachineStatus {
	return feeder.MachineStatus{
		KW:     s.KW,
		KVAR:   s.KVAR,
		Hz:     s.Hz,
		Volt:   s.Volt,
		Online: s.Online,
	}
}

// Control maps feeder.Control to feeder.DeviceControl
func mapControl(c feeder.MachineControl) Control {
	return Control{
		CloseFeeder: c.CloseFeeder,
	}
}

// LinkToBus recieves a channel from the virtual bus, which the bus will transmit its status on.
// the method returns a channel for the virtual asset to report its status to the bus.
func (a *VirtualFeeder) LinkToBus(busIn <-chan asset.VirtualStatus) <-chan asset.VirtualStatus {
	busOut := make(chan asset.VirtualStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	if err := a.Stop(); err != nil {
		panic(err)
	}

	a.startProcess()
	return busOut
}

func (a *VirtualFeeder) startProcess() {
	a.comm.recieve = make(chan Status)
	a.comm.send = make(chan Control)

	go Process(a.pid, a.comm, a.bus)
}

// Stop the virtual machine loop by closing it's communication channels.
func (a *VirtualFeeder) Stop() error {
	if a.comm.send != nil {
		//log.Println("[VirtualFeeder-Device] Stopping")
		close(a.comm.send)
	}
	return nil
}

// Process is the virtual hardware update loop
func Process(pid uuid.UUID, comm virtualHardware, bus virtualBus) {
	defer close(bus.send)
	target := &Target{pid: pid}
	sm := &stateMachine{offState{}}
	var ok bool
	log.Println("[VirtualFeeder-Device] Starting")
loop:
	for {
		select {
		case target.control, ok = <-comm.send: // write to 'hardware'
			if !ok {
				break loop
			}

		case comm.recieve <- target.status: // read from 'hardware'

		case busStatus, ok := <-bus.recieve: // read from 'virtual system'
			if !ok {
				break loop
			}
			target.status = sm.run(*target, busStatus)

		case bus.send <- target: // write to 'virtual system'

		default:
			// TODO: understand buffered/unbuffered channels in select statement...
			// These channels are all unbuffered, default seems to provide a path for execution.
			// If this isn't included, the process locks.
			time.Sleep(200 * time.Millisecond)
		}
	}
	log.Println("[VirtualFeeder-Device] Stopped")
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(target Target, bus asset.VirtualStatus) Status {
	s.currentState = s.currentState.transition(target, bus)
	return s.currentState.action(target, bus)
}

type state interface {
	action(Target, asset.VirtualStatus) Status
	transition(Target, asset.VirtualStatus) state
}

type offState struct{}

func (s offState) action(target Target, bus asset.VirtualStatus) Status {
	return Status{
		KW:     0,
		KVAR:   0,
		Hz:     bus.Hz(),
		Volt:   bus.Volt(),
		Online: false,
	}
}
func (s offState) transition(target Target, bus asset.VirtualStatus) state {
	if target.control.CloseFeeder == true {
		log.Printf("VirtualFeeder-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(target Target, bus asset.VirtualStatus) Status {
	var kw float64
	var kvar float64
	if true {
		kw = 0   //TODO: Link to a virtual load?
		kvar = 0 //TODO: Link to a virtual load?
	}
	return Status{
		KW:     kw,
		KVAR:   kvar,
		Hz:     bus.Hz(),
		Volt:   bus.Volt(),
		Online: true,
	}
}

func (s onState) transition(target Target, bus asset.VirtualStatus) state {
	if target.control.CloseFeeder == false {
		log.Printf("VirtualFeeder-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
