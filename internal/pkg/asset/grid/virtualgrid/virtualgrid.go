package virtualgrid

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualGrid target
type VirtualGrid struct {
	pid       uuid.UUID
	status    Status
	control   Control
	config    Config
	comm      Comm
	observers Observers
}

// Status data structure for the VirtualGrid
type Status struct {
	timestamp            int64
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volt                 float64 `json:"Volt"`
	PositiveRealCapacity float64 `json:"PositiveRealCapacity"`
	NegativeRealCapacity float64 `json:"NegativeRealCapacity"`
	Synchronized         bool    `json:"Synchronized"`
	Online               bool    `json:"Online"`
}

// Control data structure for the VirtualGrid
type Control struct {
	closeIntertie bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Bus string `json:"Bus"`
}

// Comm data structure for the VirtualGrid
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

type Observers struct {
	assetObserver chan<- virtualacbus.Source
	busObserver   <-chan virtualacbus.Source
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualGrid) ReadDeviceStatus(setAssetStatus func(grid.Status)) {
	a.status = a.read()
	setAssetStatus(mapStatus(a.status))
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualGrid) WriteDeviceControl(c grid.Control) {
	a.control = mapControl(c)
	a.write()
}

// Status maps grid.DeviceStatus to grid.Status
func mapStatus(s Status) grid.Status {
	// map deviceStatus to GridStatus
	return grid.Status{
		Timestamp:            s.timestamp,
		KW:                   s.KW,
		KVAR:                 s.KVAR,
		PositiveRealCapacity: s.PositiveRealCapacity,
		NegativeRealCapacity: s.NegativeRealCapacity,
		Synchronized:         s.Synchronized,
		Online:               s.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func mapControl(c grid.Control) Control {
	// map GridControl params to deviceControl
	return Control{
		closeIntertie: c.CloseIntertie,
	}
}

func (a VirtualGrid) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus := <-a.comm.incoming
	readStatus.timestamp = timestamp
	return readStatus
}

func (a VirtualGrid) write() {
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	a.comm.outgoing <- a.control
}

// New returns an initalized virtualbus Asset; this is part of the Asset interface.
func New(configPath string, buses map[string]bus.Bus) (grid.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return grid.Asset{}, err
	}

	config := Config{}
	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return grid.Asset{}, err
	}

	bus := buses[config.Bus].(virtualacbus.VirtualACBus)

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	pid, _ := uuid.NewUUID()

	device := VirtualGrid{
		pid: pid,
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
		config: config,
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
	return grid.New(jsonConfig, device)
}

func virtualDeviceLoop(comm Comm, obs Observers) {
	dev := &VirtualGrid{}
	sm := &stateMachine{offState{}}
	var ok bool
loop:
	for {
		select {
		case dev.control, ok = <-comm.outgoing:
			if !ok {
				break loop
			}
		case comm.incoming <- dev.status:
			dev.updateObservers(obs)
			dev.status = sm.run(*dev)
			//log.Printf("[VirtualGrid-Device: state: %v]\n",
			//	reflect.TypeOf(sm.currentState).String())
		}
	}
	log.Println("[VirtualGrid-Device] shutdown")
}

func (a *VirtualGrid) updateObservers(obs Observers) {
	obs.assetObserver <- a.asSource()
	if a.status.Online {
		gridformer := <-obs.busObserver
		a.status.KW = gridformer.KW
		a.status.KVAR = gridformer.KVAR
	}
}

func (a VirtualGrid) asSource() virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: a.status.Online,
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
