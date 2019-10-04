package virtualfeeder

import (
	"io/ioutil"
	"log"
	"math/rand"
	"time"

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

func (a VirtualFeeder) ReadDeviceStatus(setParentStatus func(feeder.Status)) {
	a.status = a.read()
	setParentStatus(mapStatus(a.status)) // callback for to write parent status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualFeeder) WriteDeviceControl(c feeder.Control) {
	a.control = mapControl(c)
	a.write()
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
func mapControl(c feeder.Control) Control {
	return Control{
		closeFeeder: c.CloseFeeder,
	}
}

func (a *VirtualFeeder) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus := <-a.comm.incoming
	readStatus.timestamp = timestamp
	return readStatus
}

func (a *VirtualFeeder) write() {
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	a.comm.outgoing <- a.control
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
			//log.Printf("[VirtualFeeder-Device: state: %v]\n",
			//	reflect.TypeOf(sm.currentState).String())
		}
	}
	log.Println("[VirtualFeeder-Device] shutdown")
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
