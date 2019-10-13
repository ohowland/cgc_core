package virtualfeeder

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualFeeder target
type VirtualFeeder struct {
	pid       uuid.UUID
	status    Status
	control   Control
	config    Config
	comm      Comm
	observers Observers
}

// Status data structure for the VirtualFeeder
type Status struct {
	timestamp int64
	KW        float64 `json:"KW"`
	KVAR      float64 `json:"KVAR"`
	Hz        float64 `json:"Hz"`
	Volt      float64 `json:"Volts"`
	Online    bool    `json:"Online"`
}

// Control data structure for the VirtualFeeder
type Control struct {
	closeFeeder bool
}

// StaticConfig is a data structure representing an architypical fixed feeder configuration
type Config struct { // Should this get transfered over to the specific class?
	Bus         bus.Bus
	AverageKW   float64
	AverageKVAR float64
}

// Comm data structure for the VirtualFeeder
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

type Observers struct {
	assetObserver chan<- virtualacbus.Source
	busObserver   <-chan virtualacbus.Source
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualFeeder) ReadDeviceStatus(setAssetStatus func(feeder.Status)) {
	a.status = a.read()
	setAssetStatus(mapStatus(a.status)) // callback for to write parent status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualFeeder) WriteDeviceControl(c feeder.MachineControl) {
	a.control = mapControl(c)
	a.write()
}

func (a *VirtualFeeder) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus := <-a.comm.incoming
	readStatus.timestamp = timestamp
	return readStatus
}

func (a VirtualFeeder) write() {
	a.comm.outgoing <- a.control
}

func (a *VirtualFeeder) updateObservers(obs Observers) {
	obs.assetObserver <- mapSource(*a)
}

// New returns an initalized VirtualFeeder Asset; this is part of the Asset interface.
func New(configPath string, buses map[string]bus.Bus) (feeder.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return feeder.Asset{}, err
	}

	config := struct {
		Bus         string  `json:"Bus"`
		AverageKW   float64 `json:"AverageKW"`
		AverageKVAR float64 `json:"AverageKVAR"`
	}{"", 0, 0}

	err = json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return feeder.Asset{}, err
	}

	bus := buses[config.Bus].(*virtualacbus.VirtualACBus)

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	pid, err := uuid.NewUUID()

	device := VirtualFeeder{
		pid: pid,
		status: Status{
			timestamp: 0,
			KW:        0,
			KVAR:      0,
			Hz:        0,
			Volt:      0,
			Online:    false,
		},
		control: Control{
			closeFeeder: false,
		},
		config: Config{
			Bus:         bus,
			AverageKW:   config.AverageKW,
			AverageKVAR: config.AverageKVAR,
		},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
	}

	device.startVirtualDeviceLoop()
	return feeder.New(jsonConfig, device)
}

// Status maps feeder.DeviceStatus to feeder.Status
func mapStatus(s Status) feeder.Status {
	return feeder.Status{
		Timestamp: s.timestamp,
		KW:        s.KW,
		KVAR:      s.KVAR,
		Hz:        s.Hz,
		Volt:      s.Volt,
		Online:    s.Online,
	}
}

// Control maps feeder.Control to feeder.DeviceControl
func mapControl(c feeder.MachineControl) Control {
	return Control{
		closeFeeder: c.CloseFeeder,
	}
}

func mapSource(a VirtualFeeder) virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW * -1,
		KVAR:        a.status.KVAR,
		Gridforming: false,
	}
}

func (a *VirtualFeeder) startVirtualDeviceLoop() {
	bus := a.config.Bus.(*virtualacbus.VirtualACBus)
	observers := Observers{
		assetObserver: bus.AssetObserver(),
	}

	go virtualDeviceLoop(*a, observers)
}

func (a VirtualFeeder) stopVirtualDeviceLoop() {
	close(a.observers.assetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(dev VirtualFeeder, obs Observers) {
	defer close(dev.comm.incoming)
	sm := &stateMachine{offState{}}
	var ok bool
loop:
	for {
		select {
		case dev.control, ok = <-dev.comm.outgoing:
			if !ok {
				break loop
			}
		case dev.comm.incoming <- dev.status:
			dev.updateObservers(obs)
			dev.status = sm.run(dev)
		}
	}
	log.Println("[VirtualFeeder-Device] shutdown")
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
		Hz:     dev.config.Bus.Hz(),
		Volt:   dev.config.Bus.Volt(),
		Online: false,
	}
}
func (s offState) transition(dev VirtualFeeder) state {
	if dev.control.closeFeeder == true {
		log.Printf("[VirtualFeeder-Device: state: %v]\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualFeeder) Status {
	// TODO: determine why bus pointer isn't working here.
	// unable to successfully call dev.config.Bus.Energized()
	var kw float64
	var kvar float64
	if true {
		kw = dev.config.AverageKW
		kvar = dev.config.AverageKVAR
	}
	return Status{
		KW:     kw,   // TODO: Link to virtual system model
		KVAR:   kvar, // TODO: Link to virtual system model
		Hz:     dev.config.Bus.Hz(),
		Volt:   dev.config.Bus.Volt(),
		Online: true,
	}
}

func (s onState) transition(dev VirtualFeeder) state {
	if dev.control.closeFeeder == false {
		log.Printf("[VirtualFeeder-Device: state: %v]\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
