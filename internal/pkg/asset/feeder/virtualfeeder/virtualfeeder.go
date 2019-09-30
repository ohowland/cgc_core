package virtualfeeder

import (
	"io/ioutil"
	"log"
	"reflect"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualFeeder target
type VirtualFeeder struct {
	pid       uuid.UUID
	status    Status
	control   Control
	comm      Comm
	observers Observers
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
	closeFeeder bool
}

type Config struct{}

// Comm data structure for the VirtualFeeder
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

type Observers struct {
	assetObserver chan<- virtualacbus.Source
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualFeeder) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualFeeder) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps feeder.DeviceStatus to feeder.Status
func (a VirtualFeeder) Status() feeder.Status {
	return feeder.Status{
		KW:     a.status.KW,
		KVAR:   a.status.KVAR,
		Hz:     a.status.Hz,
		Volt:   a.status.Volt,
		Online: a.status.Online,
	}
}

// Control maps feeder.Control to feeder.DeviceControl
func (a VirtualFeeder) Control(c feeder.Control) {

	updatedControl := Control{
		closeFeeder: c.CloseFeeder,
	}

	a.control = updatedControl
}

func (a *VirtualFeeder) read() error {
	select {
	case in := <-a.comm.incoming:
		a.status = in
		//log.Printf("[VirtualFeeder: read status: %v]", in)
	default:
		log.Println("[VirtualFeeder: read failed]")
	}
	return nil
}

func (a *VirtualFeeder) write() error {
	select {
	case a.comm.outgoing <- a.control:
		//log.Printf("[VirtualFeeder: write control: %v]", a.control)
	default:
		log.Println("[VirtualFeeder: write failed]")

	}
	return nil
}

// New returns an initalized VirtualFeeder Asset; this is part of the Asset interface.
func New(configPath string, bus virtualacbus.VirtualACBus) (feeder.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return feeder.Asset{}, err
	}

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	pid, err := uuid.NewUUID()

	device := VirtualFeeder{
		pid: pid,
		status: Status{
			KW:     0,
			KVAR:   0,
			Hz:     0,
			Volt:   0,
			Online: false,
		},
		control: Control{
			closeFeeder: false,
		},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
		observers: Observers{
			assetObserver: bus.AssetObserver(),
		},
	}

	go virtualDeviceLoop(device.comm, device.observers)
	return feeder.New(jsonConfig, device)
}

func virtualDeviceLoop(comm Comm, obs Observers) {
	dev := &VirtualFeeder{}
	sm := &stateMachine{offState{}}
	var ok bool
	for {
		select {
		case dev.control = <-comm.outgoing:
			if !ok {
				break
			}
		case comm.incoming <- dev.status:
			dev.updateObservers(obs)
			dev.status = sm.run(*dev)
			log.Printf("[VirtualFeeder-Device: state: %v]\n",
				reflect.TypeOf(sm.currentState).String())
		}
	}
}

func (a *VirtualFeeder) updateObservers(obs Observers) {
	obs.assetObserver <- a.asSource()
}

func (a *VirtualFeeder) asSource() virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: false,
	}
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(dev VirtualFeeder) Status {
	s.currentState = s.currentState.transition(dev)
	return s.currentState.action(dev)
}

type state interface {
	action(VirtualFeeder) Status
	transition(VirtualFeeder) state
}

type offState struct{}

func (s offState) action(dev VirtualFeeder) Status {
	return Status{
		KW:     0,
		KVAR:   0,
		Online: false,
	}
}
func (s offState) transition(dev VirtualFeeder) state {
	if dev.control.closeFeeder == true {
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualFeeder) Status {
	return Status{
		KW:     dev.status.KW,   // TODO: Link to virtual system model
		KVAR:   dev.status.KVAR, // TODO: Link to virtual system model
		Online: true,
	}
}

func (s onState) transition(dev VirtualFeeder) state {
	if dev.control.closeFeeder == false {
		return offState{}
	}
	return onState{}
}
