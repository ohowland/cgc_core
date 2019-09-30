package virtualgrid

import (
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/bus/virtualbus"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
)

// VirtualGrid target
type VirtualGrid struct {
	id      uuid.UUID
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualGrid
type Status struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volts                float64 `json:"Volts"`
	PositiveRealCapacity float64 `json:"PositiveRealCapacity"`
	NegativeRealCapacity float64 `json:"NegativeRealCapacity"`
	Synchronized         bool    `json:"Synchronized"`
	Online               bool    `json:"Online"`
}

// Control data structure for the VirtualGrid
type Control struct {
	closeIntertie bool
}

// Comm data structure for the VirtualGrid
type Comm struct {
	incoming  chan Status
	outgoing  chan Control
	busInput  chan<- virtualbus.Source
	busOutput <-chan virtualbus.Source
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualGrid) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualGrid) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a VirtualGrid) Status() grid.Status {
	// map deviceStatus to GridStatus
	return grid.Status{
		KW:                   a.status.KW,
		KVAR:                 a.status.KVAR,
		PositiveRealCapacity: a.status.PositiveRealCapacity,
		NegativeRealCapacity: a.status.NegativeRealCapacity,
		Synchronized:         a.status.Synchronized,
		Online:               a.status.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a VirtualGrid) Control(c grid.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		closeIntertie: c.CloseIntertie,
	}

	a.control = updatedControl
}

func (a *VirtualGrid) read() error {
	select {
	case in := <-a.comm.incoming:
		a.status = in
		//log.Printf("[VirtualGrid: read status: %v]", in)
	default:
		log.Println("[VirtualGrid: read failed]")
	}
	return nil
}

func (a *VirtualGrid) write() error {
	select {
	case a.comm.outgoing <- a.control:
		//log.Printf("[VirtualGrid: write control: %v]", a.control)
	default:
		log.Println("[VirtualGrid: write failed]")

	}
	return nil
}

// New returns an initalized virtualbus Asset; this is part of the Asset interface.
func New(configPath string, bus *virtualbus.VirtualBus) (grid.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return grid.Asset{}, err
	}

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	id, _ := uuid.NewUUID()

	device := VirtualGrid{
		id: id,
		status: Status{
			KW:                   0,
			KVAR:                 0,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
			Synchronized:         false,
			Online:               false,
		},
		control: Control{
			closeIntertie: false,
		},
		comm: Comm{
			incoming:  in,
			outgoing:  out,
			busInput:  bus.LoadsChan(),
			busOutput: bus.GridformChan(),
		},
	}

	go virtualDeviceLoop(device.comm)
	return grid.New(jsonConfig, device)
}

func virtualDeviceLoop(comm Comm) {
	dev := &VirtualGrid{}
	sm := &stateMachine{offState{}}
	for {
		select {
		case dev.control = <-comm.outgoing:
		case comm.incoming <- dev.status:
			dev = updateVirtualBus(dev, comm)
			dev.status = sm.run(*dev)
			log.Printf("[VirtualGrid-Device: state: %v]\n",
				reflect.TypeOf(sm.currentState).String())
		default:
			time.Sleep(time.Duration(200) * time.Millisecond)
		}
	}
}

func updateVirtualBus(dev *VirtualGrid, comm Comm) *VirtualGrid {
	dev.getGridformingLoadFromBus(comm)
	dev.reportLoadToBus(comm)
	return dev
}

func (a *VirtualGrid) getGridformingLoadFromBus(comm Comm) {
	if a.status.Online {
		select {
		case v := <-comm.busOutput:
			a.status.KW = v.KW
			a.status.KVAR = v.KVAR
		default:
		}
	}
}

func (a *VirtualGrid) reportLoadToBus(comm Comm) {
	assetLoad := virtualbus.Source{
		ID:          a.id,
		Hz:          a.status.Hz,
		Volts:       a.status.Volts,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: a.status.Online,
	}
	select {
	case comm.busInput <- assetLoad:
	default:
	}
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(dev VirtualGrid) Status {
	s.currentState = s.currentState.transition(dev)
	return s.currentState.action(dev)
}

type state interface {
	action(VirtualGrid) Status
	transition(VirtualGrid) state
}

type offState struct{}

func (s offState) action(dev VirtualGrid) Status {
	return Status{
		KW:                   0,
		KVAR:                 0,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualGrid
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualGrid
		Synchronized:         false,
		Online:               false,
	}
}
func (s offState) transition(dev VirtualGrid) state {
	if dev.control.closeIntertie == true {
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualGrid) Status {
	return Status{
		KW:                   dev.status.KW,   // TODO: Link to virtual system model
		KVAR:                 dev.status.KVAR, // TODO: Link to virtual system model
		PositiveRealCapacity: 10,              // TODO: Get Config into VirtualGrid
		NegativeRealCapacity: 10,              // TODO: Get Config into VirtualGrid
		Synchronized:         true,
		Online:               true,
	}
}

func (s onState) transition(dev VirtualGrid) state {
	if dev.control.closeIntertie == false {
		return offState{}
	}
	return onState{}
}
