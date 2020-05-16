package pv

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus() (MachineStatus, error)
	WriteDeviceControl(MachineControl) error
}

// Asset is a datastructure for an PV Asset
type Asset struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	device      DeviceController
	broadcast   map[uuid.UUID]chan<- msg.Msg
	supervisory SupervisoryControl
	config      Config
}

// PID is a getter for the unique identifier field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

func (a *Asset) Subscribe(pid uuid.UUID) <-chan msg.Msg {
	ch := make(chan msg.Msg, 1)
	a.mux.Lock()
	defer a.mux.Unlock()
	a.broadcast[pid] = ch
	return ch
}

func (a *Asset) Unsubscribe(pid uuid.UUID) {
	a.mux.Lock()
	defer a.mux.Unlock()
	ch := a.broadcast[pid]
	delete(a.broadcast, pid)
	close(ch)
}

// UpdateStatus requests a physical device read, then updates MachineStatus field.
func (a Asset) UpdateStatus() {
	machineStatus, err := a.device.ReadDeviceStatus()
	if err != nil {
		// comm fail handling path
		return
	}
	status := transform(machineStatus)
	a.mux.Lock()
	defer a.mux.Unlock()
	for _, ch := range a.broadcast {
		select {
		case ch <- msg.New(a.PID(), status):
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
		panic(errors.New("bad cast to write control"))
	}
	err := a.device.WriteDeviceControl(control)
	if err != nil {
		// comm fail handling path
	}
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) Config() Config {
	return a.config
}

func (a Asset) Enable(b bool) {
	a.supervisory.enable = b
}

// Status wraps MachineStatus with a mutex
type Status struct {
	calc    CalculatedStatus
	machine MachineStatus
}
type CalculatedStatus struct{}

// MachineStatus is a data structure representing an architypical PV status
type MachineStatus struct {
	KW     float64
	KVAR   float64
	Hz     float64
	Volt   float64
	Online bool
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
	return 0.0
}

// RealNegativeCapacity returns the asset's operative real negative capacity
func (s Status) RealNegativeCapacity() float64 {
	return 0.0
}

// MachineControl defines the hardware control interface for the ESS Asset
type MachineControl struct {
	Run     bool
	KWLimit float64
	KVAR    float64
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

type MachineConfig struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
}

// Name is a getter for the asset Name
func (c Config) Name() string {
	return c.machine.Name
}

// Bus is a getter for the asset's connected Bus
func (c Config) Bus() string {
	return c.machine.Bus
}

// New returns a configured PV Asset
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

	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}

	return Asset{&sync.Mutex{}, PID, device, broadcast, supervisory, config}, err
}
