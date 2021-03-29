package virtualpv

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/asset/pv"
)

// VirtualPV target
type VirtualPV struct {
	pid  uuid.UUID
	comm virtualHardware
	bus  virtualBus
}

// Comm data structure for the VirtualPV
type virtualHardware struct {
	send    chan Control
	recieve chan Status
}

type virtualBus struct {
	send    chan<- asset.VirtualACStatus
	recieve <-chan asset.VirtualACStatus
}

// Target is a virtual representation of the hardware
type Target struct {
	pid     uuid.UUID
	status  Status
	control Control
}

func (t Target) KW() float64 {
	return t.status.KW
}
func (t Target) KVAR() float64 {
	return t.status.KVAR

}
func (t Target) Hz() float64 {
	return t.status.Hz
}

func (t Target) Volts() float64 {
	return t.status.Volts
}

func (t Target) Gridforming() bool {
	return false
}

// Status data structure for the VirtualPV
type Status struct {
	KW     float64 `json:"KW"`
	KVAR   float64 `json:"KVAR"`
	Hz     float64 `json:"Hz"`
	Volts  float64 `json:"Volts"`
	Online bool    `json:"Online"`
}

// Control data structure for the VirtualPV
type Control struct {
	Run     bool    `json:"Run"`
	KWLimit float64 `json:"KWLimit`
	KVAR    float64 `json:"KVAR"`
}

func (a VirtualPV) PID() uuid.UUID {
	return a.pid
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualPV) ReadDeviceStatus() (pv.MachineStatus, error) {
	status, err := a.read()
	return mapStatus(status), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualPV) WriteDeviceControl(machineControl pv.MachineControl) error {
	control := mapControl(machineControl)
	err := a.write(control)
	return err
}

func (a VirtualPV) read() (Status, error) {
	fuzzing := rand.Intn(500)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus, ok := <-a.comm.recieve
	if !ok {
		return Status{}, errors.New("read error")
	}
	return readStatus, nil
}

func (a VirtualPV) write(control Control) error {
	a.comm.send <- control
	return nil
}

// New returns an initalized VirtualPV Asset; this is part of the Asset interface.
func New(configPath string) (pv.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return pv.Asset{}, err
	}

	pid, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	device := VirtualPV{
		pid:  pid,
		comm: virtualHardware{},
	}

	return pv.New(jsonConfig, &device)
}

// Status maps grid.DeviceStatus to grid.Status
func mapStatus(s Status) pv.MachineStatus {
	// map deviceStatus to GridStatus
	return pv.MachineStatus{
		KW:     s.KW,
		KVAR:   s.KVAR,
		Volts:  s.Volts,
		Hz:     s.Hz,
		Online: s.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func mapControl(c pv.MachineControl) Control {
	// map GridControl params to deviceControl
	return Control{
		Run:     c.Run,
		KWLimit: c.KWLimit,
		KVAR:    c.KVAR,
	}
}

func (a *VirtualPV) LinkToBus(busIn <-chan asset.VirtualACStatus) <-chan asset.VirtualACStatus {
	busOut := make(chan asset.VirtualACStatus)
	a.bus.send = busOut
	a.bus.recieve = busIn

	if err := a.Stop; err != nil {
		panic(err)
	}

	a.startProcess()
	return busOut
}

func (a *VirtualPV) startProcess() {
	a.comm.recieve = make(chan Status)
	a.comm.send = make(chan Control)

	go Process(a.pid, a.comm, a.bus)
}

// StopProcess stops the virtual machine loop by closing it's communication channels.
func (a *VirtualPV) Stop() error {
	if a.comm.send != nil {
		close(a.comm.send)
	}
	return nil
}

func Process(pid uuid.UUID, comm virtualHardware, bus virtualBus) {
	defer close(comm.send)
	target := &Target{pid: pid}
	sm := &stateMachine{offState{}}
	var ok bool
	log.Println("[VirtualPV-Device] Starting")
loop:
	for {
		select {
		case target.control, ok = <-comm.send: // write to 'hardware'
			if !ok {
				break loop
			}

		case comm.recieve <- target.status:

		case busStatus, ok := <-bus.recieve:
			if !ok {
				break loop
			}
			target.status = sm.run(*target, busStatus)

		case bus.send <- target:

		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
	log.Println("[VirtualPV-Device] Stopped")
}

type stateMachine struct {
	currentState state
}

func (s *stateMachine) run(target Target, bus asset.VirtualACStatus) Status {
	s.currentState = s.currentState.transition(target, bus)
	return s.currentState.action(target, bus)
}

type state interface {
	action(Target, asset.VirtualACStatus) Status
	transition(Target, asset.VirtualACStatus) state
}

func energized(bus asset.VirtualACStatus) bool {
	return bus.Hz() > 1 && bus.Volts() > 1
}

type offState struct{}

func (s offState) action(target Target, bus asset.VirtualACStatus) Status {
	return Status{
		KW:     0,
		KVAR:   0,
		Hz:     bus.Hz(),
		Volts:  bus.Volts(),
		Online: false,
	}
}

func (s offState) transition(target Target, bus asset.VirtualACStatus) state {
	if target.control.Run && energized(bus) {
		log.Printf("VirtualPV-Device: state: %v\n",
			reflect.TypeOf(onState{}).String())
		return onState{}
	}
	return offState{}
}

type onState struct{}

func (s onState) action(target Target, bus asset.VirtualACStatus) Status {

	//radiation := TotalIrradiance(a, l, time.Now())

	return Status{
		KW:     0,
		KVAR:   target.control.KVAR,
		Hz:     bus.Hz(),
		Volts:  bus.Volts(),
		Online: true,
	}
}

func (s onState) transition(target Target, bus asset.VirtualACStatus) state {
	if target.control.Run == false || !energized(bus) {
		log.Printf("VirtualPV-Device: state: %v\n",
			reflect.TypeOf(offState{}).String())
		return offState{}
	}
	return onState{}
}
