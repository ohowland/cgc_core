package virtualess

import (
	"io/ioutil"
	"log"
	"reflect"
	"sync"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualESS target
type VirtualESS struct {
	mutex     *sync.Mutex
	pid       uuid.UUID
	status    Status
	control   Control
	comm      Comm
	observers Observers
}

// Status data structure for the VirtualESS
type Status struct {
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
	run      bool
	kW       float64
	kVAR     float64
	gridForm bool
}

type Config struct {
}

// Comm data structure for the VirtualESS
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

type Observers struct {
	assetObserver chan<- virtualacbus.Source
	busObserver   <-chan virtualacbus.Source
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualESS) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualESS) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps ess.DeviceStatus to ess.Status
func (a VirtualESS) Status() ess.Status {
	// map deviceStatus to GridStatus
	return ess.Status{
		KW:                   a.status.KW,
		KVAR:                 a.status.KVAR,
		Hz:                   a.status.Hz,
		Volt:                 a.status.Volt,
		SOC:                  a.status.SOC,
		PositiveRealCapacity: a.status.PositiveRealCapacity,
		NegativeRealCapacity: a.status.NegativeRealCapacity,
		Gridforming:          a.status.Gridforming,
		Online:               a.status.Online,
	}
}

// Control maps ess.Control to ess.DeviceControl
func (a VirtualESS) Control(c ess.Control) {

	updatedControl := Control{
		run:      c.Run,
		kW:       c.KW,
		kVAR:     c.KVAR,
		gridForm: c.GridForm,
	}

	a.control = updatedControl
}

func (a *VirtualESS) read() error {
	select {
	case in := <-a.comm.incoming:
		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.status = in
		//log.Printf("[VirtualESS: read status: %v]", in)
	default:
		log.Println("[VirtualESS: read failed]")
	}
	return nil
}

func (a VirtualESS) write() error {
	select {
	case a.comm.outgoing <- a.control:
		//log.Printf("[VirtualESS: write control: %v]", a.control)
	default:
		log.Println("[VirtualESS: write failed]")

	}
	return nil
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string, bus virtualacbus.VirtualACBus) (ess.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	pid, err := uuid.NewUUID()

	device := VirtualESS{
		pid: pid,
		status: Status{
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
			run:      false,
			kW:       0,
			kVAR:     0,
			gridForm: false,
		},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
		observers: Observers{
			assetObserver: bus.AssetObserver(),
			busObserver:   bus.BusObserver(),
		},
	}

	go virtualDeviceLoop(device.comm, device.observers)
	return ess.New(jsonConfig, device)
}

func virtualDeviceLoop(comm Comm, obs Observers) {
	dev := &VirtualESS{} // The virtual 'hardware' device
	sm := &stateMachine{offState{}}
	var ok bool
	for {
		select {
		case dev.control, ok = <-comm.outgoing: // write to 'hardware'
			if !ok {
				break
			}
		case comm.incoming <- dev.status: // read from 'hardware'
			dev.updateObservers(obs)
			dev.status = sm.run(*dev)
			log.Printf("[VirtualESS-Device: state: %v]\n",
				reflect.TypeOf(sm.currentState).String())
		}
	}
	//log.Println("[VirtualESS-Device: shutdown]")
}

func (a *VirtualESS) updateObservers(obs Observers) {
	obs.assetObserver <- a.asSource()
	if a.status.Gridforming {
		gridformer := <-obs.busObserver

		a.mutex.Lock()
		defer a.mutex.Unlock()
		a.status.KW = gridformer.KW
		a.status.KVAR = gridformer.KVAR
	}
}

func (a VirtualESS) asSource() virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: a.status.Gridforming,
	}
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
	if dev.control.run == true {
		if dev.control.gridForm == true {
			return HzVState{}
		}
		return PQState{}
	}
	return offState{}
}

type PQState struct{}

func (s PQState) action(dev VirtualESS) Status {
	return Status{
		KW:                   dev.status.KW,
		KVAR:                 dev.status.KVAR,
		SOC:                  dev.status.SOC,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualESS
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualESS
		Gridforming:          false,
		Online:               true,
	}
}

func (s PQState) transition(dev VirtualESS) state {
	if dev.control.run == false {
		return offState{}
	}
	if dev.control.gridForm == true {
		return HzVState{}
	}
	return PQState{}
}

type HzVState struct{}

func (s HzVState) action(dev VirtualESS) Status {
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

func (s HzVState) transition(dev VirtualESS) state {
	if dev.control.run == false {
		return offState{}
	}
	if dev.control.gridForm == false {
		return PQState{}
	}
	return HzVState{}
}
