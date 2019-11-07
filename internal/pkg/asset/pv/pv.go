package pv

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(int64, MachineStatus))
	WriteDeviceControl(MachineControl)
}

// Asset is a datastructure for an PV Asset
type Asset struct {
	pid     uuid.UUID
	device  DeviceController
	status  Status
	control Control
	config  Config
}

// PID is a getter for the unique identifier field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.status.setStatus)
}

// WriteControl requests a physical device write of the data held in the PV control field.
func (a Asset) WriteControl() {
	a.control.mux.Lock()
	defer a.control.mux.Unlock()
	control := a.control.machine
	go a.device.WriteDeviceControl(control)
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
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

// Status wraps MachineStatus with a mutex
type Status struct {
	mux       *sync.Mutex
	timestamp int64
	machine   MachineStatus
}

// MachineStatus is a data structure representing an architypical PV status
type MachineStatus struct {
	KW     float64
	KVAR   float64
	Hz     float64
	Volt   float64
	Online bool
}

func (s *Status) setStatus(timestamp int64, ms MachineStatus) {
	if timestamp > s.timestamp { // mux before?
		s.mux.Lock()
		defer s.mux.Unlock()
		s.machine = ms
	}
}

// KW returns the asset's measured real power
func (s Status) KW() float64 {
	return s.machine.KW
}

// KVAR returns the asset's measured reactive power
func (s Status) KVAR() float64 {
	return s.machine.KVAR
}

// Control is a data structure representing an architypical PV control
type Control struct {
	mux         *sync.Mutex
	machine     MachineControl
	supervisory SupervisoryControl
}

// MachineControl defines the hardware control interface for the ESS Asset
type MachineControl struct {
	Run     bool
	KWLimit float64
	KVAR    float64
}

// supervisoryControl defines the software control interface for the ESS Asset
type SupervisoryControl struct {
	Enable bool
	Manual bool
}

// KW sets the PV Inverter's power limit
func (c *Control) KW(kw float64) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.machine.KWLimit = kw
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

	status := Status{&sync.Mutex{}, 0, MachineStatus{}}
	control := Control{
		&sync.Mutex{},
		MachineControl{false, 0, 0},
		SupervisoryControl{false, false},
	}
	config := Config{&sync.Mutex{}, machineConfig}
	return Asset{PID, device, status, control, config}, err
}
