package virtualbms

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/asset/bms"
)

// VirtualBMS target
type VirtualBMS struct {
	pid  uuid.UUID
	comm virtualHardware
	bus  virtualBus
}

// For commmunication between asset and virtual hardware
type virtualHardware struct {
	send    chan Control
	recieve chan Status
}

// For commmunication between virtual bus and asset
type virtualBus struct {
	send    chan<- asset.VirtualDCStatus
	recieve <-chan asset.VirtualDCStatus
}

// Target is a virtual representation of the hardware
type Target struct {
	pid     uuid.UUID
	status  Status
	control Control
}

// KW is an accbmsor for real power
func (t Target) KW() float64 {
	return t.status.KW
}

// Volts is an accbmsor for ac voltage
func (t Target) Volts() float64 {
	return t.status.Volts
}

func (t Target) Gridforming() bool {
	return t.status.Online
}

// Status data structure for the VirtualBMS
type Status struct {
	KW                   float64 `json:"KW"`
	Volts                float64 `json:"Volts"`
	SOC                  float64 `json:"SOC"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	Online               bool    `json:"Online"`
}

// Control data structure for the VirtualBMS
type Control struct {
	Run bool    `json:"Run"`
	KW  float64 `json:"KW"`
}

// PID is an accbmsor for the process id
func (a VirtualBMS) PID() uuid.UUID {
	return a.pid
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualBMS) ReadDeviceStatus() (bms.MachineStatus, error) {
	status, err := a.read()
	return mapStatus(status), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualBMS) WriteDeviceControl(machineControl bms.MachineControl) error {
	control := mapControl(machineControl)
	err := a.write(control)
	return err
}

func (a VirtualBMS) read() (Status, error) {
	fuzzing := rand.Intn(500)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.recieve
	if !ok {
		return Status{}, errors.New("read error")
	}
	return readStatus, nil
}

func (a VirtualBMS) write(control Control) error {
	a.comm.send <- control
	return nil
}

// New returns an initalized VirtualBMS Asset; this is part of the Asset interface.
func New(configPath string) (bms.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return bms.Asset{}, err
	}

	pid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	device := VirtualBMS{
		pid:  pid,
		comm: virtualHardware{},
	}

	return bms.New(jsonConfig, &device)
}

// Status maps bms.DeviceStatus to bms.Status
func mapStatus(s Status) bms.MachineStatus {
	return bms.MachineStatus{
		KW:                   s.KW,
		Volts:                s.Volts,
		SOC:                  s.SOC,
		RealPositiveCapacity: s.RealPositiveCapacity,
		RealNegativeCapacity: s.RealNegativeCapacity,
		Online:               s.Online,
	}
}

// Control maps bms.Control to bms.DeviceControl
func mapControl(c bms.MachineControl) Control {
	return Control{
		Run: c.Run,
		KW:  c.KW,
	}
}

// LinkToBus recieves a channel from the virtual bus, which the bus will transmit its status on.
// the method returns a channel for the virtual asset to report its status to the bus.
func (a *VirtualBMS) LinkToBus(busIn <-chan asset.VirtualDCStatus) <-chan asset.VirtualDCStatus {
	busOut := make(chan asset.VirtualDCStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	if err := a.Stop(); err != nil {
		panic(err)
	}

	a.startProcess()
	return busOut
}

// startProcbms spawns virtual hardware which the virtual bms communicates with
func (a *VirtualBMS) startProcess() {
	a.comm.recieve = make(chan Status)
	a.comm.send = make(chan Control)

	go Process(a.pid, a.comm, a.bus)
}

// Stop stops the virtual machine loop by closing it's communication channels.
func (a *VirtualBMS) Stop() error {
	if a.comm.send != nil {
		//log.Println("[VirtualBMS-Device] Stopping")
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
	log.Println("[VirtualBMS-Device] Starting")
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
	log.Println("[VirtualBMS-Device] Stopped")
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(target Target, bus asset.VirtualDCStatus) Status {
	s.currentState = s.currentState.transition(target, bus)
	return s.currentState.action(target, bus)
}

type state interface {
	action(Target, asset.VirtualDCStatus) Status
	transition(Target, asset.VirtualDCStatus) state
}

type offState struct{}

func (s offState) action(target Target, bus asset.VirtualDCStatus) Status {
	return Status{
		KW:                   0,
		Volts:                bus.Volts(),
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 0, // TODO: Get Config into VirtualBMS
		RealNegativeCapacity: 0, // TODO: Get Config into VirtualBMS
		Online:               false,
	}
}

func (s offState) transition(target Target, bus asset.VirtualDCStatus) state {
	if target.control.Run {
		log.Printf("VirtualBMS-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

// onState is the bus-connected state
type onState struct{}

func (s onState) action(target Target, bus asset.VirtualDCStatus) Status {
	return Status{
		KW:                   bus.KW(),
		Volts:                800, // TODO: Get Config into VirtualBMS
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 10, // TODO: Get Config into VirtualBMS
		RealNegativeCapacity: 10, // TODO: Get Config into VirtualBMS
		Online:               true,
	}
}

func (s onState) transition(target Target, bus asset.VirtualDCStatus) state {
	if !target.control.Run {
		log.Printf("VirtualBMS-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}

	return onState{}
}
