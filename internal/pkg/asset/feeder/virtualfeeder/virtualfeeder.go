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
	comm comm
	bus  virtualBus
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

func (t Target) KW() float64 {
	return t.status.KW
}
func (t Target) KVAR() float64 {
	return t.status.KVAR

}
func (t Target) Hz() float64 {
	return t.status.Hz
}

func (t Target) Volt() float64 {
	return t.status.Volt
}

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

// Comm data structure for the VirtualFeeder
type comm struct {
	incoming chan Status
	outgoing chan Control
}

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
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.incoming
	if !ok {
		return Status{}, errors.New("Read Error")
	}
	return readStatus, nil
}

func (a VirtualFeeder) write(control Control) error {
	a.comm.outgoing <- control
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
		comm: comm{},
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

func (a *VirtualFeeder) LinkToBus(busIn <-chan asset.VirtualStatus) <-chan asset.VirtualStatus {
	busOut := make(chan asset.VirtualStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	a.StopProcess()
	a.StartProcess()
	return busOut
}

func (a *VirtualFeeder) StartProcess() {
	in := make(chan Status)
	out := make(chan Control)
	a.comm.incoming = in
	a.comm.outgoing = out

	go Process(a.pid, a.comm, a.bus)
}

// StopProcess stops the virtual machine loop by closing it's communication channels.
func (a *VirtualFeeder) StopProcess() {
	if a.comm.outgoing != nil {
		close(a.comm.outgoing)
	}
}

func Process(pid uuid.UUID, comm comm, bus virtualBus) {
	defer close(comm.incoming)
	target := &Target{pid: pid}
	sm := &stateMachine{offState{}}
	var ok bool

loop:
	for {
		select {
		case target.control, ok = <-comm.outgoing: // write to 'hardware'
			if !ok {
				break loop
			}
		case comm.incoming <- target.status:
		case busStatus := <-bus.recieve:
			target.status = sm.run(*target, busStatus)
		case bus.send <- target:
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	log.Println("VirtualFeeder-Device shutdown")
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

func energized(bus asset.VirtualStatus) bool {
	return bus.Hz() > 1 && bus.Volt() > 1
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
		Hz:     60.0, // TODO: Link to virtual system model
		Volt:   480,
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
