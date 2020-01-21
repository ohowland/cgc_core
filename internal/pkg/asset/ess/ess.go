package ess

import (
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus() (MachineStatus, error)
	WriteDeviceControl(MachineControl) error
}

// Asset is a data structure for an ESS Asset
type Asset struct {
	mux          *sync.Mutex
	pid          uuid.UUID
	device       DeviceController
	broadcast    map[uuid.UUID]chan<- asset.Msg
	control      <-chan asset.Msg
	controlOwner uuid.UUID
	supervisory  SupervisoryControl
	config       Config
}

// PID is a getter for the asset PID
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// Subscribe returns a read only channel for the asset's status.
func (a *Asset) Subscribe(pid uuid.UUID) <-chan asset.Msg {
	ch := make(chan asset.Msg)
	a.mux.Lock()
	defer a.mux.Unlock()
	a.broadcast[pid] = ch
	return ch
}

// Unsubscribe closes the broadcast channel associated with the pid parameter.
func (a *Asset) Unsubscribe(pid uuid.UUID) {
	a.mux.Lock()
	defer a.mux.Unlock()
	if ch, ok := a.broadcast[pid]; ok {
		delete(a.broadcast, pid)
		close(ch)
	}

}

// RequestControl connects the asset control to the read only channel parameter.
func (a *Asset) RequestControl(pid uuid.UUID, ch <-chan asset.Msg) bool {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control = ch
	a.controlOwner = pid
	return true
}

// UpdateStatus requests a physical device read, then updates MachineStatus field.
func (a Asset) UpdateStatus() {
	machineStatus, err := a.device.ReadDeviceStatus()
	if err != nil {
		log.Printf("ESS: %v Comm Error\n", err)
		return
	}
	status := transform(machineStatus)
	a.mux.Lock()
	defer a.mux.Unlock()
	for _, broadcast := range a.broadcast {
		select {
		case broadcast <- asset.NewMsg(a.PID(), status):
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

// WriteControl requests a physical device write of the data held in the asset machine control field.
func (a Asset) WriteControl(c interface{}) {
	control, ok := c.(MachineControl)
	if !ok {
		panic(errors.New("ESS bad cast to write control"))
	}
	err := a.device.WriteDeviceControl(control)
	if err != nil {
		log.Printf("ESS: %v Comm Error\n", err)
	}
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) Config() Config {
	return a.config
}

// Enable is an settor for the asset enable state
func (a *Asset) Enable(b bool) {
	a.supervisory.enable = b
}

// Status wraps MachineStatus with mutex and state metadata
type Status struct {
	calc    CalculatedStatus
	machine MachineStatus
}

// CalculatedStatus is a data structure representing asset state information
// that is calculated from data read into the archetype ess.
type CalculatedStatus struct{}

// MachineStatus is a data structure representing an architypical ESS status
type MachineStatus struct {
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volt                 float64
	RealPositiveCapacity float64
	RealNegativeCapacity float64
	SOC                  float64
	Gridforming          bool
	Online               bool
}

// KW returns the asset's measured real power
func (s Status) KW() float64 {
	return s.machine.KW
}

// KVAR returns the asset's measured reactive power
func (s Status) KVAR() float64 {
	return s.machine.KVAR
}

// RealPositiveCapacity returns the asset's operative real positive capacity
func (s Status) RealPositiveCapacity() float64 {
	return s.machine.RealPositiveCapacity
}

// RealNegativeCapacity returns the asset's operative real negative capacity
func (s Status) RealNegativeCapacity() float64 {
	return s.machine.RealNegativeCapacity
}

// MachineControl defines the hardware control interface for the ESS Asset
type MachineControl struct {
	Run      bool
	KW       float64
	KVAR     float64
	Gridform bool
}

// SupervisoryControl defines the software control interface for the ESS Asset
type SupervisoryControl struct {
	mux    *sync.Mutex
	enable bool
}

// Config wraps MachineConfig with mutex a mutex and hides the internal state.
type Config struct {
	mux     *sync.Mutex
	machine MachineConfig
}

// MachineConfig holds the ESS asset configuration parameters
type MachineConfig struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
	RatedKWH  float64 `json:"RatedKWH"`
}

// Name is a getter for the asset Name
func (c Config) Name() string {
	return c.machine.Name
}

// Bus is a getter for the asset's connected Bus
func (c Config) Bus() string {
	return c.machine.Bus
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

	broadcast := make(map[uuid.UUID]chan<- asset.Msg)

	var control <-chan asset.Msg
	controlOwner := PID

	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}

	return Asset{&sync.Mutex{}, PID, device, broadcast, control, controlOwner, supervisory, config}, err
}
