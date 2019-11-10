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
	comm comm
	bus  virtualBus
}

type virtualBus struct {
	send    chan<- asset.VirtualAssetStatus
	recieve <-chan asset.VirtualAssetStatus
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

// Comm data structure for the VirtualESS
type comm struct {
	incoming chan Status
	outgoing chan Control
}

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
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.incoming
	if !ok {
		return Status{}, errors.New("Read Error")
	}
	return readStatus, nil
}

func (a VirtualESS) write(control Control) error {
	a.comm.outgoing <- control
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
		comm: comm{},
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

func (a *VirtualESS) LinkToBus(busIn <-chan asset.VirtualAssetStatus) <-chan asset.VirtualAssetStatus {
	busOut := make(chan asset.VirtualAssetStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	a.StopProcess()
	a.StartProcess()
	return busOut
}

func (a *VirtualESS) StartProcess() {
	in := make(chan Status)
	out := make(chan Control)
	a.comm.incoming = in
	a.comm.outgoing = out

	go Process(a.pid, a.comm, a.bus)
}

// StopVirtualDevice stops the virtual machine loop by closing it's communication channels.
func (a *VirtualESS) StopProcess() {
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
			target.status = sm.run(*target)
		case _, ok = <-bus.recieve:
		case bus.send <- target:
		}
	}
	log.Println("VirtualESS-Device shutdown")
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(target Target) Status {
	s.currentState = s.currentState.transition(target)
	return s.currentState.action(target)
}

type state interface {
	action(Target) Status
	transition(Target) state
}

type offState struct{}

func (s offState) action(target Target) Status {
	return Status{
		KW:                   0,
		KVAR:                 0,
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 0, // TODO: Get Config into VirtualESS
		RealNegativeCapacity: 0, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               false,
	}
}
func (s offState) transition(target Target) state {
	if target.control.Run == true {
		if target.control.Gridform == true {
			log.Printf("VirtualESS-Device: state: %v\n",
				reflect.TypeOf(hzVState{}).String())
			return hzVState{}
		}
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(pQState{}).String())
		return pQState{}
	}
	return offState{}
}

// pQState is the power control state
type pQState struct{}

func (s pQState) action(target Target) Status {
	return Status{
		KW:                   target.control.KW,
		KVAR:                 target.control.KVAR,
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 10, // TODO: Get Config into VirtualESS
		RealNegativeCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               true,
	}
}

func (s pQState) transition(target Target) state {
	if target.control.Run == false {
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

func (s hzVState) action(target Target) Status {
	return Status{
		KW:                   target.status.KW,   // TODO: Link to virtual system model
		KVAR:                 target.status.KVAR, // TODO: Link to virtual system model
		SOC:                  target.status.SOC,
		RealPositiveCapacity: 10, // TODO: Get Config into VirtualESS
		RealNegativeCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          true,
		Online:               true,
	}
}

func (s hzVState) transition(target Target) state {
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
