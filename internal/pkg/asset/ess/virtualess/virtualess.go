package virtualess

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
)

// VirtualESS target
type VirtualESS struct {
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualESS
type Status struct {
	KW                   float64
	KVAR                 float64
	SOC                  float64
	PositiveRealCapacity float64
	NegativeRealCapacity float64
	GridForming          bool
}

// Control data structure for the VirtualESS
type Control struct {
	Run      bool
	KW       float64
	KVAR     float64
	GridForm bool
}

type VirtualSystemModel struct {
	RealSwingLoad     float64
	ReactiveSwingLoad float64
}

// Comm data structure for the VirtualESS
type Comm struct {
	vsm      chan VirtualSystemModel
	incoming chan Status
	outgoing chan Control
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

// Status maps grid.DeviceStatus to grid.Status
func (a VirtualESS) Status() ess.Status {
	// map deviceStatus to GridStatus
	return ess.Status{
		KW:                   float64(a.status.KW),
		KVAR:                 float64(a.status.KVAR),
		SOC:                  float64(a.status.SOC),
		PositiveRealCapacity: float64(a.status.PositiveRealCapacity),
		NegativeRealCapacity: float64(a.status.NegativeRealCapacity),
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a VirtualESS) Control(c ess.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		Run:      c.Run,
		KW:       c.KW,
		KVAR:     c.KVAR,
		GridForm: c.GridForm,
	}

	a.control = updatedControl
}

func (a *VirtualESS) read() error {
	select {
	case in := <-a.comm.incoming:
		a.status = in
	default:
	}
	return nil
}

func (a *VirtualESS) write() error {
	select {
	case a.comm.outgoing <- a.control:
	default:
	}
	return nil
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string, vsm chan VirtualSystemModel) (ess.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	in := make(chan Status)
	out := make(chan Control)

	device := VirtualESS{
		status: Status{
			KW:                   0,
			KVAR:                 0,
			SOC:                  0.6,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
			GridForming:          false,
		},
		control: Control{
			Run:      false,
			KW:       0,
			KVAR:     0,
			GridForm: false,
		},
		comm: Comm{
			vsm,
			in,
			out,
		},
	}

	go launchVirtualDevice(device.comm)
	return ess.New(jsonConfig, device)
}

func launchVirtualDevice(comm Comm) {
	dev := &VirtualESS{}
	sm := &stateMachine{offState{}}
	for {
		select {
		case dev.control = <-comm.outgoing:
			log.Println("[VirtualESS: control msg recieved]")
		case comm.incoming <- dev.status:
			dev = updateVirtualDevice(dev, comm.vsm)
			dev.status = sm.run(*dev)
			log.Printf("[VirtualESS: state %v]", sm.currentState)
		default:
			time.Sleep(time.Duration(200) * time.Millisecond)
		}
	}
}

func updateVirtualDevice(dev *VirtualESS, vsm chan VirtualSystemModel) *VirtualESS {
	select {
	case model := <-vsm:
		dev.status.KW = model.RealSwingLoad
		dev.status.KVAR = model.ReactiveSwingLoad
	default:
	}
	return dev
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
	}
}
func (s offState) transition(dev VirtualESS) state {
	if dev.control.Run == true {
		if dev.control.GridForm == true {
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
	}
}

func (s PQState) transition(dev VirtualESS) state {
	if dev.control.Run == false {
		return offState{}
	}
	if dev.control.GridForm == true {
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
	}
}

func (s HzVState) transition(dev VirtualESS) state {
	if dev.control.Run == false {
		return offState{}
	}
	if dev.control.GridForm == false {
		return PQState{}
	}
	return HzVState{}
}
