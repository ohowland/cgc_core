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
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualESS target
type VirtualESS struct {
	pid       uuid.UUID
	status    Status
	control   Control
	comm      Comm
	observers virtualacbus.Observers
}

// Status data structure for the VirtualESS
type Status struct {
	timestamp            int64
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volt                 float64 `json:"Volt"`
	SOC                  float64 `json:"SOC"`
	PositiveRealCapacity float64 `json:"PositiveRealCapacity"`
	NegativeRealCapacity float64 `json:"NegativeRealCapacity"`
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
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a *VirtualESS) ReadDeviceStatus(setParentStatus func(ess.Status)) {
	a.status = a.read()
	setParentStatus(mapStatus(a.status)) // callback for to write parent status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualESS) WriteDeviceControl(c ess.MachineControl) {
	a.control = mapControl(c)
	a.write()
}

func (a VirtualESS) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.incoming
	if !ok {
		log.Println("Read Error: VirtualESS, virtual channel is not open")
		return Status{}
	}
	readStatus.timestamp = timestamp
	return readStatus
}

func (a VirtualESS) write() {
	a.comm.outgoing <- a.control
}

func (a *VirtualESS) updateObservers(obs virtualacbus.Observers) {
	source := mapSource(*a)
	obs.AssetObserver <- source
	if a.status.Gridforming {
		gridformer := <-obs.BusObserver
		a.status.KW = gridformer.KW
		a.status.KVAR = gridformer.KVAR
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
		pid: pid,
		status: Status{
			timestamp:            0,
			KW:                   0,
			KVAR:                 0,
			Hz:                   0,
			Volt:                 0,
			SOC:                  0.6,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
			Gridforming:          false,
			Online:               false,
		},
		control: Control{
			Run:      false,
			KW:       0,
			KVAR:     0,
			Gridform: false,
		},
		comm: Comm{},
	}

	//device.startVirtualDeviceLoop()
	return ess.New(jsonConfig, &device)
}

// Status maps ess.DeviceStatus to ess.Status
func mapStatus(s Status) ess.Status {
	// map deviceStatus to GridStatus
	return ess.Status{
		Timestamp:            s.timestamp,
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

func mapSource(a VirtualESS) virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: a.status.Gridforming,
	}
}

// LinkVirtualDevice pulls the communication channels from the virtual bus and holds them in asset.observers
func (a *VirtualESS) LinkToBus(b bus.Bus) error {
	vrACbus, ok := b.(virtualacbus.VirtualACBus)
	if !ok {
		return errors.New("Bus cannot be cast to VirtualACBus")
	}
	a.observers = vrACbus.GetBusObservers()
	return nil
}

// StartVirualDevice launches the virtual machine loop.
func (a *VirtualESS) StartVirualDevice() {
	in := make(chan Status, 1)
	out := make(chan Control, 1)
	a.comm.incoming = in
	a.comm.outgoing = out

	go virtualDeviceLoop(a.pid, a.comm, a.observers)
}

// StopVirtualDevice stops the virtual machine loop by closing it's communication channels.
func (a VirtualESS) StopVirtualDevice() {
	close(a.observers.AssetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(pid uuid.UUID, comm Comm, obs virtualacbus.Observers) {
	defer close(comm.incoming)
	dev := &VirtualESS{pid: pid} // The virtual 'hardware' device
	sm := &stateMachine{offState{}}
	var ok bool
loop:
	for {
		select {
		case dev.control, ok = <-comm.outgoing: // write to 'hardware'
			if !ok {
				break loop
			}
		case comm.incoming <- dev.status: // read from 'hardware'
			dev.updateObservers(obs)
			dev.status = sm.run(*dev)
		}
	}
	log.Println("VirtualESS-Device shutdown")
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(dev VirtualESS) Status {
	s.currentState = s.currentState.transition(dev)
	return s.currentState.action(dev)
}

type state interface {
	action(VirtualESS) Status
	transition(VirtualESS) state
}

type offState struct{}

func (s offState) action(dev VirtualESS) Status {
	return Status{
		KW:                   0,
		KVAR:                 0,
		SOC:                  dev.status.SOC,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               false,
	}
}
func (s offState) transition(dev VirtualESS) state {
	if dev.control.Run == true {
		if dev.control.Gridform == true {
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

func (s pQState) action(dev VirtualESS) Status {
	return Status{
		KW:                   dev.control.KW,
		KVAR:                 dev.control.KVAR,
		SOC:                  dev.status.SOC,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               true,
	}
}

func (s pQState) transition(dev VirtualESS) state {
	if dev.control.Run == false {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	if dev.control.Gridform == true {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(hzVState{}).String())
		return hzVState{}
	}
	return pQState{}
}

// hzVState is the gridforming state
type hzVState struct{}

func (s hzVState) action(dev VirtualESS) Status {
	return Status{
		KW:                   dev.status.KW,   // TODO: Link to virtual system model
		KVAR:                 dev.status.KVAR, // TODO: Link to virtual system model
		SOC:                  dev.status.SOC,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          true,
		Online:               true,
	}
}

func (s hzVState) transition(dev VirtualESS) state {
	if dev.control.Run == false {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	if dev.control.Gridform == false {
		log.Printf("VirtualESS-Device: state: %v\n",
			reflect.TypeOf(pQState{}).String())
		return pQState{}
	}
	return hzVState{}
}
