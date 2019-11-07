package virtualfeeder

import (
	"errors"
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
	comm      comm
	observers virtualacbus.Observers
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

// Comm data structure for the VirtualFeeder
type comm struct {
	incoming chan Status
	outgoing chan Control
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualFeeder) ReadDeviceStatus(setAssetStatus func(int64, feeder.MachineStatus)) {
	timestamp := time.Now().UnixNano()
	a.status = a.read()
	setAssetStatus(timestamp, mapStatus(a.status)) // callback for to write parent status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualFeeder) WriteDeviceControl(c feeder.MachineControl) {
	a.control = mapControl(c)
	a.write()
}

func (a *VirtualFeeder) read() Status {
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus := <-a.comm.incoming
	return readStatus
}

func (a VirtualFeeder) write() {
	a.comm.outgoing <- a.control
}

func (a *VirtualFeeder) updateObservers(obs virtualacbus.Observers) {
	source := mapSource(*a)
	obs.AssetObserver <- source
}

// New returns an initalized VirtualFeeder Asset; this is part of the Asset interface.
func New(configPath string) (feeder.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return feeder.Asset{}, err
	}

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
		comm: comm{},
	}

	return feeder.New(jsonConfig, &device)
}

// Status maps feeder.DeviceStatus to feeder.Status
func mapStatus(s Status) feeder.MachineStatus {
	return feeder.MachineStatus{
		KW:     s.KW,
		KVAR:   s.KVAR,
		Hz:     s.Hz,
		Volt:   s.Volt,
		Online: s.Online,
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

// LinkToBus pulls the communication channels from the virtual bus and holds them in asset.observers
func (a *VirtualFeeder) LinkToBus(b bus.Bus) error {
	vrACbus, ok := b.(virtualacbus.VirtualACBus)
	if !ok {
		return errors.New("Bus cannot be cast to VirtualACBus")
	}
	a.observers = vrACbus.GetBusObservers()
	return nil
}

// StartVirtualDevice launches the virtual machine loop
func (a *VirtualFeeder) StartVirtualDevice() {
	in := make(chan Status, 1)
	out := make(chan Control, 1)
	a.comm.incoming = in
	a.comm.outgoing = out
	go virtualDeviceLoop(a.pid, a.comm, a.observers)
}

// StopVirtualDevice stops the virutal machine loop
func (a VirtualFeeder) StopVirtualDevice() {
	close(a.observers.AssetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(pid uuid.UUID, comm comm, obs virtualacbus.Observers) {
	defer close(comm.incoming)
	dev := &VirtualFeeder{pid: pid} // The virtual 'hardware' device
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
	log.Println("VirtualFeeder-Device: shutdown")
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
		Hz:     0,
		Volt:   0,
		Online: false,
	}
}
func (s offState) transition(dev VirtualFeeder) state {
	if dev.control.closeFeeder == true {
		log.Printf("VirtualFeeder-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualFeeder) Status {
	var kw float64
	var kvar float64
	if true {
		kw = 456 //TODO: Link to a virtual load?
		kvar = 123
	}
	return Status{
		KW:     kw,
		KVAR:   kvar,
		Hz:     60.0, // TODO: Link to virtual system model
		Volt:   480,
		Online: true,
	}
}

func (s onState) transition(dev VirtualFeeder) state {
	if dev.control.closeFeeder == false {
		log.Printf("VirtualFeeder-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
