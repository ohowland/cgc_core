package grid

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus() (MachineStatus, error)
	WriteDeviceControl(MachineControl) error
}

// Asset is a datastructure for an Energy Storage System Asset
type Asset struct {
	mux          *sync.Mutex
	pid          uuid.UUID
	device       DeviceController
	broadcast    map[uuid.UUID]chan<- msg.Msg
	controlOwner uuid.UUID
	supervisory  SupervisoryControl
	config       Config
}

// PID is a getter for the asset PID
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Name is a getter for the asset Name
func (a Asset) Name() string {
	return a.config.machine.Name
}

// Bus is a getter for the asset's connected Bus
func (a Asset) Bus() string {
	return a.config.machine.Bus
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

func (a *Asset) Subscribe(pid uuid.UUID) <-chan msg.Msg {
	ch := make(chan msg.Msg)
	a.mux.Lock()
	defer a.mux.Unlock()
	a.broadcast[pid] = ch
	return ch
}

func (a *Asset) Unsubscribe(pid uuid.UUID) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if ch, ok := a.broadcast[pid]; ok {
		delete(a.broadcast, pid)
		close(ch)
	}
}

// RequestControl connects the asset control to the read only channel parameter.
func (a *Asset) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) bool {
	a.mux.Lock()
	defer a.mux.Unlock()
	// TODO: previous owner needs to stop. how to enforce?
	a.controlOwner = pid
	go a.controlHandler(ch)

	return true
}

// UpdateStatus requests a physical device read, then updates MachineStatus field.
func (a Asset) UpdateStatus() {
	machineStatus, err := a.device.ReadDeviceStatus()
	if err != nil {
		log.Printf("Grid: %v Comm Error\n", err)
		return
	}
	status := transform(machineStatus)
	a.mux.Lock()
	defer a.mux.Unlock()
	for _, broadcast := range a.broadcast {
		select {
		case broadcast <- msg.New(a.PID(), msg.STATUS, status):
		default:
		}
	}
}

func transform(machineStatus MachineStatus) Status {
	return Status{
		CalculatedStatus{},
		machineStatus,
	}
}

func (a *Asset) controlHandler(ch <-chan msg.Msg) {
loop:
	for {
		msg, ok := <-ch
		if !ok {
			log.Println("Feeder controlHandler() stopping")
			break loop
		}

		control, ok := msg.Payload().(MachineControl)
		if !ok {
			log.Println("Feeder controlHandler() bad type assertion")
		}

		err := a.device.WriteDeviceControl(control)
		if err != nil {
			log.Println("Feeder controlHandler():", err)
		}
	}
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) Config() Config {
	return a.config
}

// Enable is an settor for the asset enable state
func (a Asset) Enable(b bool) {
	a.supervisory.enable = b
}

// Status wraps MachineStatus with mutex and state metadata
type Status struct {
	Calc    CalculatedStatus `json:"CalculatedStatus"`
	Machine MachineStatus    `json:"MachineStatus"`
}

// CalculatedStatus is a data structure representing asset state information
// that is calculated from data read into the archetype grid.
type CalculatedStatus struct{}

// MachineStatus is a data structure representing an architypical Grid Intertie status
type MachineStatus struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volts                float64 `json:"Volts"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	Online               bool    `json:"Online"`
}

// KW returns the asset's measured real power
func (s Status) KW() float64 {
	return s.Machine.KW
}

// KVAR returns the asset's measured reactive power
func (s Status) KVAR() float64 {
	return s.Machine.KVAR
}

// RealPositiveCapacity returns the asset's operative real positive capacity
func (s Status) RealPositiveCapacity() float64 {
	return s.Machine.RealPositiveCapacity
}

// RealNegativeCapacity returns the asset's operative real negative capacity
func (s Status) RealNegativeCapacity() float64 {
	return s.Machine.RealNegativeCapacity
}

// MachineControl represents the control state of the machine
type MachineControl struct {
	CloseIntertie bool
}

type SupervisoryControl struct {
	mux    *sync.Mutex
	enable bool
}

// Config wraps MachineConfig with mutex a mutex and hides the internal state.
type Config struct {
	mux     *sync.Mutex
	machine MachineConfig
}

type MachineConfig struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
}

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	machineConfig := MachineConfig{}
	err := json.Unmarshal(jsonConfig, &machineConfig)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	broadcast := make(map[uuid.UUID]chan<- msg.Msg)

	controlOwner := PID

	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}

	return Asset{&sync.Mutex{}, PID, device, broadcast, controlOwner, supervisory, config}, err

}
