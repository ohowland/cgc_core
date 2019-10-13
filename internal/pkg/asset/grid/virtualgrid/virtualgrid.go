package virtualgrid

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
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
	Bus bus.Bus
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
func (a VirtualGrid) WriteDeviceControl(c grid.MachineControl) {
	a.control = mapControl(c)
	a.write()
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
	a.comm.outgoing <- a.control
}

func (a *VirtualGrid) updateObservers(obs Observers) {
	obs.assetObserver <- mapSource(*a)
	if a.status.Online {
		gridformer := <-obs.busObserver
		a.status.KW = gridformer.KW
		a.status.KVAR = gridformer.KVAR
	}
}

// New returns an initalized virtualbus Asset; this is part of the Asset interface.
func New(configPath string, buses map[string]bus.Bus) (grid.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return grid.Asset{}, err
	}

	config := struct {
		Bus string `json:"Bus"`
	}{""}

	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return grid.Asset{}, err
	}

	bus := buses[config.Bus].(*virtualacbus.VirtualACBus)

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
		config: Config{
			Bus: bus,
		},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
	}

	device.startVirtualDeviceLoop()
	return grid.New(jsonConfig, device)
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
func mapControl(c grid.MachineControl) Control {
	// map GridControl params to deviceControl
	return Control{
		closeIntertie: c.CloseIntertie,
	}
}

func mapSource(a VirtualGrid) virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: a.status.Online,
	}
}

func (a *VirtualGrid) startVirtualDeviceLoop() {
	bus := a.config.Bus.(*virtualacbus.VirtualACBus)
	observers := Observers{
		assetObserver: bus.AssetObserver(),
		busObserver:   bus.BusObserver(),
	}

	go virtualDeviceLoop(a.pid, a.comm, observers)
}

func (a VirtualGrid) stopVirtualDeviceLoop() {
	close(a.observers.assetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(pid uuid.UUID, comm Comm, obs Observers) {
	defer close(comm.incoming)
	dev := &VirtualGrid{pid: pid}
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
		}
	}
	log.Println("[VirtualGrid-Device] shutdown")
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
		Hz:                   0,
		Volt:                 0,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualGrid
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualGrid
		Synchronized:         true,
		Online:               false,
	}
}
func (s offState) transition(dev VirtualGrid) state {
	if dev.control.closeIntertie == true {
		log.Printf("VirtualGrid-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualGrid) Status {
	return Status{
		KW:                   dev.status.KW,   // TODO: Link to virtual system model
		KVAR:                 dev.status.KVAR, // TODO: Link to virtual system model
		Hz:                   60,
		Volt:                 480,
		PositiveRealCapacity: 10, // TODO: Get Config into VirtualGrid
		NegativeRealCapacity: 10, // TODO: Get Config into VirtualGrid
		Synchronized:         true,
		Online:               true,
	}
}

func (s onState) transition(dev VirtualGrid) state {
	if dev.control.closeIntertie == false {
		log.Printf("VirtualGrid-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
