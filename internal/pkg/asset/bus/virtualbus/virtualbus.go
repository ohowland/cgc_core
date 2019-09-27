package virtualbus

import (
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/virtual"

	"github.com/ohowland/cgc/internal/pkg/asset/bus"
)

const (
	queueSize = 50
)

// VirtualBus target
type VirtualBus struct {
	id      uuid.UUID
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualBus
type Status struct {
	Hz    float64               `json:"Hz"`
	Volts float64               `json:"Volts"`
	Loads map[uuid.UUID]float64 `json:"Loads"`
}

// Control data structure for the VirtualBus
type Control struct{}

type Config struct{}

// Comm data structure for the VirtualBus
type Comm struct {
	incoming chan Status
	outgoing chan Control
	loadsIn  chan virtual.SourceLoad
	loadsOut chan virtual.Load
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualBus) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualBus) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps bus.DeviceStatus to bus.Status
func (a VirtualBus) Status() bus.Status {
	return bus.Status{
		Hz:        float64(a.status.Hz),
		Volts:     float64(a.status.Volts),
		Energized: false,
	}
}

// Control maps bus.Control to bus.DeviceControl
func (a VirtualBus) Control(c bus.Control) {

	updatedControl := Control{}

	a.control = updatedControl
}

func (a *VirtualBus) read() error {
	select {
	case in := <-a.comm.incoming:
		a.status = in
		//log.Printf("[VirtualBus: read status: %v]", in)
	default:
		log.Println("[VirtualBus: read failed]")
	}
	return nil
}

func (a *VirtualBus) write() error {
	select {
	case a.comm.outgoing <- a.control:
		//log.Printf("[VirtualBus: write control: %v]", a.control)
	default:
		log.Println("[VirtualBus: write failed]")

	}
	return nil
}

// New returns an initalized VirtualBus Asset; this is part of the Asset interface.
func New(configPath string, vsm *virtual.SystemModel) (bus.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return bus.Asset{}, err
	}

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	loadsIn := make(chan virtual.SourceLoad, queueSize)
	loadsOut := make(chan virtual.Load, 1)

	id, err := uuid.NewUUID()

	device := VirtualBus{
		id: id,
		status: Status{
			Hz:    0.0,
			Volts: 0.0,
		},
		control: Control{},
		comm: Comm{
			incoming: in,
			outgoing: out,
			loadsIn:  vsm.ReportLoad,
			loadsOut: vsm.SwingLoad,
		},
	}

	go launchVirtualDevice(device.comm)
	return bus.New(jsonConfig, device)
}

func launchVirtualDevice(comm Comm) {
	dev := &VirtualBus{}
	sm := &stateMachine{offState{}}
	for {
		select {
		case dev.control = <-comm.outgoing:
		case comm.incoming <- dev.status:
			dev = updateVirtualSystem(dev, comm)
			dev.status = sm.run(*dev)
			log.Printf("[VirtualBus-Device: state: %v]\n",
				reflect.TypeOf(sm.currentState).String())
		default:
			time.Sleep(time.Duration(200) * time.Millisecond)
		}
	}
}

func updateVirtualSystem(dev *VirtualBus, comm Comm) *VirtualBus {
	select {
	case asset := <-comm.loadsIn:
		dev.status.loads[asset.ID] = asset.Load
		//log.Printf("[VirtualSystemModel: Reported Load %v]\n", v)
	case s.SwingLoad <- s.calcSwingLoad():
		log.Printf("[VirtualSystemModel: Swing Load %v]\n", s.calcSwingLoad())
	}
	return dev
}

func (s SystemModel) calcSwingLoad() Load {
	kwSum := 0.0
	kvarSum := 0.0
	for _, l := range s.loads {
		kwSum += l.KW
		kvarSum += l.KVAR
	}
	return Load{
		KW:   kwSum,
		KVAR: kvarSum,
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
