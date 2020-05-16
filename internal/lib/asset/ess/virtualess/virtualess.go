package virtualess

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/ess"
)

// VirtualESS target
type VirtualESS struct {
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
	return t.status.Gridforming
}

// Status data structure for the VirtualESS
type Status struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volt                 float64 `json:"Volt"`
	SOC                  float64 `json:"SOC"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	Gridforming          bool    `json:"Gridforming"`
	Online               bool    `json:"Online"`
}

// Control data structure for the VirtualESS
type Control struct {
	Run      bool    `json:"Run"`
	KW       float64 `json:"KW"`
	KVAR     float64 `json:"KVAR"`
	Gridform bool    `json:"Gridform"`
}

// PID is an accessor for the process id
func (a VirtualESS) PID() uuid.UUID {
	return a.pid
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualESS) ReadDeviceStatus() (ess.MachineStatus, error) {
	status, err := a.read()
	return mapStatus(status), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualESS) WriteDeviceControl(machineControl ess.MachineControl) error {
	control := mapControl(machineControl)
	err := a.write(control)
	return err
}

func (a VirtualESS) read() (Status, error) {
	fuzzing := rand.Intn(500)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.recieve
	if !ok {
		return Status{}, errors.New("Read Error")
	}
	return readStatus, nil
}

func (a VirtualESS) write(control Control) error {
	a.comm.send <- control
	return nil
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string) (ess.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	pid, err := uuid.NewUUID()

	device := VirtualESS{
		pid:  pid,
		comm: virtualHardware{},
	}

	return ess.New(jsonConfig, &device)
}

// Status maps ess.DeviceStatus to ess.Status
func mapStatus(s Status) ess.MachineStatus {
	return ess.MachineStatus{
		KW:                   s.KW,
		KVAR:                 s.KVAR,
		Hz:                   s.Hz,
		Volt:                 s.Volt,
		SOC:                  s.SOC,
		RealPositiveCapacity: s.RealPositiveCapacity,
		RealNegativeCapacity: s.RealNegativeCapacity,
		Gridforming:          s.Gridforming,
		Online:               s.Online,
	}
}

// Control maps ess.Control to ess.DeviceControl
func mapControl(c ess.MachineControl) Control {
	return Control{
		Run:      c.Run,
		KW:       c.KW,
		KVAR:     c.KVAR,
		Gridform: c.Gridform,
	}
}

// LinkToBus recieves a channel from the virtual bus, which the bus will transmit its status on.
// the method returns a channel for the virtual asset to report its status to the bus.
func (a *VirtualESS) LinkToBus(busIn <-chan asset.VirtualStatus) <-chan asset.VirtualStatus {
	busOut := make(chan asset.VirtualStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	if err := a.Stop(); err != nil {
		panic(err)
	}

	a.startProcess()
	return busOut
}

// startProcess spawns virtual hardware which the virtual ess communicates with
func (a *VirtualESS) startProcess() {
	a.comm.recieve = make(chan Status)
	a.comm.send = make(chan Control)

	go Process(a.pid, a.comm, a.bus)
}

// Stop stops the virtual machine loop by closing it's communication channels.
func (a *VirtualESS) Stop() error {
	if a.comm.send != nil {
		//log.Println("[VirtualESS-Device] Stopping")
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
	log.Println("[VirtualESS-Device] Starting")
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
	log.Println("[VirtualESS-Device] Stopped")
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
		KW:                   0,
		KVAR:                 0,
		Hz:                   bus.Hz(),
		Volt:                 bus.Volt(),
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 0, // TODO: Get Config into VirtualESS
		RealNegativeCapacity: 0, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               false,
	}
}

func (s offState) transition(target Target, bus asset.VirtualStatus) state {
	if target.control.Run == true {
		if target.control.Gridform == true {
			log.Printf("VirtualESS-Device: state: %v\n",
				reflect.TypeOf(hzVState{}).String())
			return hzVState{}
		}

		if bus.Gridforming() {
			log.Printf("VirtualESS-Device: state: %v\n",
				reflect.TypeOf(pQState{}).String())
			return pQState{}
		}
	}
	return offState{}
}

// pQState is the power control state
type pQState struct{}

func (s pQState) action(target Target, bus asset.VirtualStatus) Status {
	return Status{
		KW:                   target.control.KW,
		KVAR:                 target.control.KVAR,
		Hz:                   bus.Hz(),
		Volt:                 bus.Volt(),
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 10, // TODO: Get Config into VirtualESS
		RealNegativeCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               true,
	}
}

func (s pQState) transition(target Target, bus asset.VirtualStatus) state {
	if target.control.Run == false || !bus.Gridforming() {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	if target.control.Gridform == true {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(hzVState{}).String())
		return hzVState{}
	}
	return pQState{}
}

// hzVState is the gridforming state
type hzVState struct{}

func (s hzVState) action(target Target, bus asset.VirtualStatus) Status {
	return Status{
		KW:                   bus.KW(),
		KVAR:                 bus.KVAR(),
		Hz:                   60,  // TODO: Get Config into VirtualESS
		Volt:                 480, // TODO: Get Config into VirtualESS
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 10, // TODO: Get Config into VirtualESS
		RealNegativeCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          true,
		Online:               true,
	}
}

func (s hzVState) transition(target Target, bus asset.VirtualStatus) state {
	if target.control.Run == false {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	if target.control.Gridform == false {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(pQState{}).String())
		return pQState{}
	}
	return hzVState{}
}
