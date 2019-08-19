package virtualess

import (
	"io/ioutil"

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
	KW                   int `json:"KW"`
	KVAR                 int `json:"KVAR"`
	SOC                  int `json:"SOC"`
	PositiveRealCapacity int `json:"PositiveRealCapacity"`
	NegativeRealCapacity int `json:"NegativeRealCapacity"`
}

// Control data structure for the VirtualESS
type Control struct {
	runRequest bool
}

// Comm data structure for the VirtualESS
type Comm struct {
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
		runRequest: c.RunRequest,
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

// New returns an initalized IPC30C3 Asset; this is part of the Asset interface.
func New(configPath string) (ess.Asset, error) {
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
			SOC:                  0,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
		},
		control: Control{
			runRequest: false,
		},
		comm: Comm{in, out},
	}

	ess, err := ess.New(jsonConfig, device)
	go launchVirtualDevice(out, in)
	return ess, err
}

func launchVirtualDevice(in chan Control, out chan Status) {
	dev := &VirtualESS{}
	sm := &stateMachine{offState{}}
	for {
		select {
		case dev.control = <-in:
		case out <- dev.status:
		default:
			dev.status = sm.run(*dev)
		}
	}
}

type state interface {
	action(VirtualESS) Status
	transition(VirtualESS) state
}

type offState struct{}

func (s offState) action(dev VirtualESS) Status {
	return Status{}
}
func (s offState) transition(dev VirtualESS) state {
	if dev.control.runRequest == true {
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualESS) Status {
	return Status{}
}
func (s onState) transition(dev VirtualESS) state {
	if dev.control.runRequest == false {
		return offState{}
	}
	return onState{}
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(dev VirtualESS) Status {
	s.currentState = s.currentState.transition(dev)
	return s.currentState.action(dev)
}
