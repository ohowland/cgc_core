package virtualgrid

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/asset/grid"
)

// VirtualGrid target
type VirtualGrid struct {
	pid  uuid.UUID
	comm virtualHardware
	bus  virtualBus
}

// Comm data structure for the VirtualGrid
type virtualHardware struct {
	send    chan Control
	recieve chan Status
}

type virtualBus struct {
	send    chan<- asset.VirtualACStatus
	recieve <-chan asset.VirtualACStatus
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

// Volts is an accessor for ac voltage
func (t Target) Volts() float64 {
	return t.status.Volts
}

// Gridforming is an accessor for gridforming state
func (t Target) Gridforming() bool {
	return t.status.Online
}

// Status data structure for the VirtualGrid
type Status struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volts                float64 `json:"Volts"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	Online               bool    `json:"Online"`
}

// Control data structure for the VirtualGrid
type Control struct {
	CloseIntertie bool
}

// PID is an accessor for the process id
func (a VirtualGrid) PID() uuid.UUID {
	return a.pid
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualGrid) ReadDeviceStatus() (grid.MachineStatus, error) {
	status, err := a.read()
	return mapStatus(status), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualGrid) WriteDeviceControl(machineControl grid.MachineControl) error {
	control := mapControl(machineControl)
	err := a.write(control)
	return err
}

func (a VirtualGrid) read() (Status, error) {
	fuzzing := rand.Intn(500)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.recieve
	if !ok {
		return Status{}, errors.New("Read Error")
	}
	return readStatus, nil
}

func (a VirtualGrid) write(control Control) error {
	a.comm.send <- control
	return nil
}

// New returns an initalized virtualbus Asset; this is part of the Asset interface.
func New(configPath string) (grid.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return grid.Asset{}, err
	}

	pid, _ := uuid.NewUUID()

	device := VirtualGrid{
		pid:  pid,
		comm: virtualHardware{},
	}

	return grid.New(jsonConfig, &device)
}

// Status maps grid.DeviceStatus to grid.Status
func mapStatus(s Status) grid.MachineStatus {
	// map deviceStatus to GridStatus
	return grid.MachineStatus{
		KW:                   s.KW,
		KVAR:                 s.KVAR,
		Hz:                   s.Hz,
		Volts:                s.Volts,
		RealPositiveCapacity: s.RealPositiveCapacity,
		RealNegativeCapacity: s.RealNegativeCapacity,
		Online:               s.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func mapControl(c grid.MachineControl) Control {
	// map GridControl params to deviceControl
	return Control{
		CloseIntertie: c.CloseIntertie,
	}
}

// LinkToBus recieves a channel from the virtual bus, which the bus will transmit its status on.
// the method returns a channel for the virtual asset to report its status to the bus.
func (a *VirtualGrid) LinkToBus(busIn <-chan asset.VirtualACStatus) <-chan asset.VirtualACStatus {
	busOut := make(chan asset.VirtualACStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	if err := a.Stop(); err != nil {
		panic(err)
	}
	a.startProcess()
	return busOut
}

func (a *VirtualGrid) startProcess() {
	a.comm.recieve = make(chan Status)
	a.comm.send = make(chan Control)

	go Process(a.pid, a.comm, a.bus)
}

// Stop the virtual machine loop by closing it's communication channels.
func (a *VirtualGrid) Stop() error {
	if a.comm.send != nil {
		//log.Println("[VirtualGrid-Device] Stopping")
		close(a.comm.send)
	}
	return nil
}

// Process is an asynchronous routine representing the hardware device.
func Process(pid uuid.UUID, comm virtualHardware, bus virtualBus) {
	defer close(bus.send)
	target := &Target{pid: pid}
	sm := &stateMachine{offState{}}
	var ok bool
	log.Println("[VirtualGrid-Device] Starting")
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
			time.Sleep(200 * time.Millisecond)
		}
	}
	log.Println("[VirtualGrid-Device] Stopped")
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(target Target, bus asset.VirtualACStatus) Status {
	s.currentState = s.currentState.transition(target, bus)
	return s.currentState.action(target, bus)
}

type state interface {
	action(Target, asset.VirtualACStatus) Status
	transition(Target, asset.VirtualACStatus) state
}

type offState struct{}

func (s offState) action(target Target, bus asset.VirtualACStatus) Status {
	return Status{
		KW:                   0,
		KVAR:                 0,
		Hz:                   bus.Hz(),    // TODO: Get Config into VirtualGrid
		Volts:                bus.Volts(), // TODO: Get Config into VirtualGrid
		RealPositiveCapacity: 0,           // TODO: Get Config into VirtualGrid
		RealNegativeCapacity: 0,           // TODO: Get Config into VirtualGrid
		Online:               false,
	}
}
func (s offState) transition(target Target, bus asset.VirtualACStatus) state {
	if target.control.CloseIntertie == true {
		log.Printf("VirtualGrid-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(target Target, bus asset.VirtualACStatus) Status {
	return Status{
		KW:                   bus.KW(),
		KVAR:                 bus.KVAR(),
		Hz:                   60,  // TODO: Get Config into VirtualGrid
		Volts:                480, // TODO: Get Config into VirtualGrid
		RealPositiveCapacity: 10,  // TODO: Get Config into VirtualGrid
		RealNegativeCapacity: 10,  // TODO: Get Config into VirtualGrid
		Online:               true,
	}
}

func (s onState) transition(target Target, bus asset.VirtualACStatus) state {
	if target.control.CloseIntertie == false {
		log.Printf("VirtualGrid-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
