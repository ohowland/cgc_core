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
	Hz               float64              `json:"Hz"`
	Volts            float64              `json:"Volts"`
	ConnectedSources map[uuid.UUID]Source `json:"Loads"`
}

// Control data structure for the VirtualBus
type Control struct{}

type Config struct{}

// Comm data structure for the VirtualBus
type Comm struct {
	incoming      chan Status
	outgoing      chan Control
	sourcesIn     chan Source
	gridformerOut chan Source
}

type Source struct {
	ID          uuid.UUID
	Voltage     float64
	Frequency   float64
	KW          float64
	KVAR        float64
	Gridforming bool
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

	in := make(chan Status, 1)
	out := make(chan Control, 1)

	sourcesIn := make(chan Source, queueSize)
	gridformerOut := make(chan Source, 1)

	id, err := uuid.NewUUID()

	device := VirtualBus{
		id: id,
		status: Status{
			Hz:    0.0,
			Volts: 0.0,
		},
		control: Control{},
		comm: Comm{
			incoming:      in,
			outgoing:      out,
			sourcesIn:     sourcesIn,
			gridformerOut: gridformerOut,
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
		default:
			dev = updateVirtualSystem(dev, comm) // TODO: calc every loop?
			dev.status = sm.run(*dev)
			log.Printf("[VirtualBus-SystemModel: state: %v]\n",
				reflect.TypeOf(sm.currentState).String())
			time.Sleep(time.Duration(200) * time.Millisecond)
		}
	}
}

func updateVirtualSystem(dev *VirtualBus, comm Comm) *VirtualBus {
	select {
	case s := <-comm.sourcesIn:
		dev.status.ConnectedSources[s.ID] = s
		//log.Printf("[VirtualBus-SystemModel: Reported Load %v]\n", v)
	case comm.gridformerOut <- dev.swingMachineLoad():
		log.Printf("[VirtualBus-SystemModel: Swing Load %v]\n",
			dev.swingMachineLoad())
	}
	return dev
}

func (a VirtualBus) swingMachineLoad() Source {
	kwSum := 0.0
	kvarSum := 0.0
	var swingMachine Source
	for _, s := range a.status.ConnectedSources {
		if s.Gridforming != true {
			kwSum += s.KW
			kvarSum += s.KVAR
		} else {
			swingMachine = s
		}
	}

	swingMachine.KW = kwSum
	swingMachine.KVAR = kvarSum
	return swingMachine
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(dev VirtualBus) Status {
	s.currentState = s.currentState.transition(dev)
	return s.currentState.action(dev)
}

type state interface {
	action(VirtualBus) Status
	transition(VirtualBus) state
}

type offState struct{}

func (s offState) action(dev VirtualBus) Status {
	return Status{
		Hz:    0,
		Volts: 0,
	}
}
func (s offState) transition(dev VirtualBus) state {
	if dev.control.closeFeeder == true {
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualBus) Status {
	return Status{
		Hz:        dev.status.KW,   // TODO: Link to virtual system model
		Volts:     dev.status.KVAR, // TODO: Link to virtual system model	}
}

func (s onState) transition(dev VirtualBus) state {
	if dev.control.closeFeeder == false {
		return offState{}
	}
	return onState{}
}
