package virtualess

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
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
	config    Config
	comm      Comm
	observers Observers
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
	run      bool
	kW       float64
	kVAR     float64
	gridForm bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	bus bus.Bus
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
func (a VirtualESS) ReadDeviceStatus(setParentStatus func(ess.Status)) {
	a.status = a.read()
	setParentStatus(mapStatus(a.status)) // callback for to write parent status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualESS) WriteDeviceControl(c ess.Control) {
	a.control = mapControl(c)
	a.write()
}

func (a VirtualESS) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus := <-a.comm.incoming
	readStatus.timestamp = timestamp
	return readStatus
}

func (a VirtualESS) write() {
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	a.comm.outgoing <- a.control
}

func (a *VirtualESS) updateObservers(obs Observers) {
	obs.assetObserver <- mapSource(*a)
	if a.status.Gridforming {
		gridformer := <-obs.busObserver
		a.status.KW = gridformer.KW
		a.status.KVAR = gridformer.KVAR
	}
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string, buses map[string]bus.Bus) (ess.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	config := struct {
		Bus string `json:"Bus"`
	}{""}

	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return ess.Asset{}, err
	}

	bus := buses[config.Bus].(*virtualacbus.VirtualACBus)

	in := make(chan Status, 1)
	out := make(chan Control, 1)

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
			run:      false,
			kW:       0,
			kVAR:     0,
			gridForm: false,
		},
		config: Config{
			bus: bus,
		},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
	}

	device.startVirtualDeviceLoop()
	return ess.New(jsonConfig, device)
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
func mapControl(c ess.Control) Control {
	return Control{
		run:      c.Run,
		kW:       c.KW,
		kVAR:     c.KVAR,
		gridForm: c.GridForm,
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

func (a *VirtualESS) startVirtualDeviceLoop() {

	bus := a.config.bus.(*virtualacbus.VirtualACBus)
	observers := Observers{
		assetObserver: bus.AssetObserver(),
		busObserver:   bus.BusObserver(),
	}

	go virtualDeviceLoop(a.comm, observers)
}

func (a VirtualESS) stopVirtualDeviceLoop() {
	close(a.observers.assetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(comm Comm, obs Observers) {
	defer close(comm.incoming)
	dev := &VirtualESS{} // The virtual 'hardware' device
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
			//log.Printf("[VirtualESS-Device: state: %v]\n",
			//	reflect.TypeOf(sm.currentState).String())
		}
	}
	log.Println("[VirtualESS-Device] shutdown")
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
