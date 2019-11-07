package grid

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

// Asset is a datastructure for an Energy Storage System Asset
type Asset struct {
	pid     uuid.UUID
	device  DeviceController
	status  Status
	control Control
	config  Config
}

// PID is a getter for the asset PID
func (a Asset) PID() uuid.UUID {
	return a.pid
}

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

type Status struct {
	mux       *sync.Mutex
	timestamp int64
	machine   MachineStatus
}

// MachineStatus is a data structure representing an architypical Grid Intertie status
type MachineStatus struct {
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volts                float64
	PositiveRealCapacity float64
	NegativeRealCapacity float64
	Synchronized         bool
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

func (s *Status) setStatus(timestamp int64, ms MachineStatus) {
	if timestamp > s.timestamp { // mux before?
		s.mux.Lock()
		defer s.mux.Unlock()
		s.machine = ms
	}
}

// Control is a data structure representing an architypical Grid Intertie control
type Control struct {
	mux         *sync.Mutex
	machine     MachineControl
	supervisory SupervisoryControl
}

// MachineControl represents the control state of the machine
type MachineControl struct {
	CloseIntertie bool
}

type SupervisoryControl struct {
	Enable bool
	Manual bool
}

// KW sets the asset's real power setpoint
func (c *Control) KW(kw float64) {}

// KVAR sets the asset's reactive power setpoint
func (c *Control) KVAR(kvar float64) {}

// Run sets the asset's run request state
func (c *Control) Run(run bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.machine.CloseIntertie = run
}

// Gridform sets the asset's gridform request state
func (c *Control) Gridform(gridform bool) {}

// Config differentiates between two types of configurations, static and dynamic
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
		MachineControl{false},
		SupervisoryControl{false, false},
	}
	config := Config{&sync.Mutex{}, machineConfig}
	return Asset{PID, device, status, control, config}, err

}
