package ess

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for an ESS Asset
type Asset struct {
	pid     uuid.UUID
	device  DeviceController
	status  Status
	control Control
	config  Config
}

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(int64, MachineStatus))
	WriteDeviceControl(MachineControl)
}

// PID is a getter for the asset PID
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// UpdateStatus requests a physical device read, then updates MachineStatus field.
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.status.setStatus)
}

// WriteControl requests a physical device write of the data held in the GridAsset control field.
func (a Asset) WriteControl() {
	a.control.mux.Lock()
	defer a.control.mux.Unlock()
	control := a.control.machine
	go a.device.WriteDeviceControl(control)
}

// Status returns the archetypical status for the energy storage system asset.
// This takes the form of the ess.MachineStatus struct
func (a Asset) Status() Status {
	return a.status
}

// Control returns a pointer to the machine control struct.
func (a *Asset) Control() *Control {
	return &a.control
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) Config() Config {
	return a.config
}

// Status wraps MachineStatus with mutex and state metadata
type Status struct {
	mux       *sync.Mutex
	timestamp int64
	machine   MachineStatus
}

// MachineStatus is a data structure representing an architypical ESS status
type MachineStatus struct {
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volt                 float64
	PositiveRealCapacity float64
	NegativeRealCapacity float64
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

func (s *Status) setStatus(timestamp int64, updatedStatus MachineStatus) {
	if timestamp > s.timestamp { // mux before?
		s.mux.Lock()
		defer s.mux.Unlock()
		s.machine = updatedStatus
	}
}

// Control wraps MachineControl and SupervisoryControl with mutex
type Control struct {
	mux         *sync.Mutex
	machine     MachineControl
	supervisory SupervisoryControl
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
	Enable bool
	Manual bool
}

// KW sets the asset's real power setpoint
func (c *Control) KW(kw float64) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.machine.KW = kw
}

// KVAR sets the asset's reactive power setpoint
func (c *Control) KVAR(kvar float64) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.machine.KVAR = kvar
}

// Run sets the asset's run request state
func (c *Control) Run(run bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.machine.Run = run
}

// Gridform sets the asset's gridform request state
func (c *Control) Gridform(gridform bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.machine.Gridform = gridform
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

	status := Status{&sync.Mutex{}, 0, MachineStatus{}}
	control := Control{
		&sync.Mutex{},
		MachineControl{false, 0, 0, false},
		SupervisoryControl{false, false},
	}
	config := Config{&sync.Mutex{}, machineConfig}
	return Asset{PID, device, status, control, config}, err
}
