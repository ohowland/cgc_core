package virtualess

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus/virtualacbus"
)

// VirtualESS target
type VirtualESS struct {
	pid       uuid.UUID
	comm      comm
	observers virtualacbus.Observers
}

// Target is a virtual representation of the hardware
type Target struct {
	pid     uuid.UUID
	status  Status
	control Control
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

func (t *Target) updateObservers(obs virtualacbus.Observers) {
	source := mapSource(*t)
	obs.AssetObserver <- source
	if t.status.Gridforming {
		gridformer := <-obs.BusObserver
		t.status.KW = gridformer.KW
		t.status.KVAR = gridformer.KVAR
	}
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
		PositiveRealCapacity: s.PositiveRealCapacity,
		NegativeRealCapacity: s.NegativeRealCapacity,
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

func mapSource(t Target) virtualacbus.Source {
	return virtualacbus.Source{
		PID:         t.pid,
		Hz:          t.status.Hz,
		Volt:        t.status.Volt,
		KW:          t.status.KW,
		KVAR:        t.status.KVAR,
		Gridforming: t.status.Gridforming,
	}
}

// LinkToBus pulls the communication channels from the virtual bus and holds them in asset.observers
func (a *VirtualESS) LinkToBus(b bus.Bus) error {
	vrACbus, ok := b.(virtualacbus.VirtualACBus)
	if !ok {
		return errors.New("Bus cannot be cast to VirtualACBus")
	}
	a.observers = vrACbus.GetBusObservers()
	return nil
}

// StartVirtualDevice launches the virtual machine loop.
func (a *VirtualESS) StartVirtualDevice() {
	in := make(chan Status)
	out := make(chan Control)
	a.comm.incoming = in
	a.comm.outgoing = out
	go virtualDeviceLoop(a.pid, a.comm, a.observers)
}

// StopVirtualDevice stops the virtual machine loop by closing it's communication channels.
func (a VirtualESS) StopVirtualDevice() {
	close(a.observers.AssetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(pid uuid.UUID, comm comm, obs virtualacbus.Observers) {
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
			target.updateObservers(obs)
			target.status = sm.run(*target)
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
		PositiveRealCapacity: 0, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 0, // TODO: Get Config into VirtualESS
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
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualESS
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
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualESS
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
