package virtualpv

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/pv"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualPV target
type VirtualPV struct {
	pid       uuid.UUID
	status    Status
	control   Control
	comm      comm
	observers virtualacbus.Observers
}

// Status data structure for the VirtualPV
type Status struct {
	timestamp int64
	KW        float64 `json:"KW"`
	KVAR      float64 `json:"KVAR"`
	Hz        float64 `json:"Hz"`
	Volt      float64 `json:"Volt"`
	Online    bool    `json:"Online"`
}

// Control data structure for the VirtualPV
type Control struct {
	Run bool `json:"Run"`
}

// Config is a data structure representing an architypical fixed PV configuration
type Config struct {
	Bus string `json:"Bus"`
}

// Comm data structure for the VirtualPV
type comm struct {
	incoming chan Status
	outgoing chan Control
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a *VirtualPV) ReadDeviceStatus(setAssetStatus func(pv.Status)) {
	a.status = a.read()
	setAssetStatus(mapStatus(a.status)) // callback for to write archetype status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualPV) WriteDeviceControl(c pv.MachineControl) {
	a.control = mapControl(c)
	a.write()
}

func (a VirtualPV) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.incoming
	if !ok {
		log.Println("Read Error: VirtualESS, virtual channel is not open")
		return Status{}
	}
	readStatus.timestamp = timestamp
	return readStatus
}

func (a VirtualPV) write() {
	a.comm.outgoing <- a.control
}

func (a *VirtualPV) updateObservers(obs virtualacbus.Observers) {
	source := mapSource(*a)
	obs.AssetObserver <- source
}

// New returns an initalized VirtualPV Asset; this is part of the Asset interface.
func New(configPath string) (pv.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return pv.Asset{}, err
	}

	pid, err := uuid.NewUUID()

	device := VirtualPV{
		pid: pid,
		status: Status{
			KW:     0,
			KVAR:   0,
			Hz:     0,
			Volt:   0,
			Online: false,
		},
		control: Control{
			Run: false,
		},
		comm: comm{},
	}

	return pv.New(jsonConfig, &device)
}

// Status maps grid.DeviceStatus to grid.Status
func mapStatus(s Status) pv.Status {
	// map deviceStatus to GridStatus
	return pv.Status{
		Timestamp: s.timestamp,
		KW:        s.KW,
		KVAR:      s.KVAR,
		Volt:      s.Volt,
		Hz:        s.Hz,
		Online:    s.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func mapControl(c pv.MachineControl) Control {
	// map GridControl params to deviceControl
	return Control{
		Run: c.Run,
	}
}

func mapSource(a VirtualPV) virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: false,
	}
}

// LinkToBus pulls the communication channels from the virtual bus and holds them in asset.observers
func (a *VirtualPV) LinkToBus(b bus.Bus) error {
	vrACbus, ok := b.(virtualacbus.VirtualACBus)
	if !ok {
		return errors.New("Bus cannot be cast to VirtualACBus")
	}
	a.observers = vrACbus.GetBusObservers()
	return nil
}

// StartVirtualDevice launches the virtual machine loop.
func (a *VirtualPV) StartVirtualDevice() {
	in := make(chan Status, 1)
	out := make(chan Control, 1)
	a.comm.incoming = in
	a.comm.outgoing = out

	go virtualDeviceLoop(a.pid, a.comm, a.observers)
}

// StopVirtualDevice stops the virtual machine loop by closing it's communication channels.
func (a VirtualPV) StopVirtualDevice() {
	close(a.observers.AssetObserver)
	close(a.comm.outgoing)
}

func virtualDeviceLoop(pid uuid.UUID, comm comm, obs virtualacbus.Observers) {
	defer close(comm.incoming)
	dev := &VirtualPV{} // The virtual 'hardware' device
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
		}
	}
	log.Println("[VirtualPV-Device] shutdown")

}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(dev VirtualPV) Status {
	s.currentState = s.currentState.transition(dev)
	return s.currentState.action(dev)
}

type state interface {
	action(VirtualPV) Status
	transition(VirtualPV) state
}

type offState struct{}

func (s offState) action(dev VirtualPV) Status {
	return Status{
		KW:     0,
		KVAR:   0,
		Hz:     0,
		Volt:   0,
		Online: false,
	}
}

func (s offState) transition(dev VirtualPV) state {
	if dev.control.Run == true {
		log.Printf("VirtualPV-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(dev VirtualPV) Status {
	return Status{
		KW:     0,
		KVAR:   0,
		Hz:     0,
		Volt:   0,
		Online: false,
	}
}

func (s onState) transition(dev VirtualPV) state {
	if dev.control.Run == false {
		log.Printf("VirtualPV-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
