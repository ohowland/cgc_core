package ess

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for an ESS Asset
type Asset struct {
	mux     *sync.Mutex
	pid     uuid.UUID
	device  DeviceController
	status  Status
	control Control
	config  Config
}

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(Status))
	WriteDeviceControl(MachineControl)
}

// Status is a data structure representing an architypical ESS status
type Status struct {
	Timestamp            int64
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volt                 float64
	SOC                  float64
	PositiveRealCapacity float64
	NegativeRealCapacity float64
	Gridforming          bool
	Online               bool
}

// Control holds the ESS asset control parameters
type Control struct {
	machine     MachineControl
	supervisory supervisoryControl
}

// MachineControl defines the hardware control interface for the ESS Asset
type MachineControl struct {
	Run      bool
	KW       float64
	KVAR     float64
	Gridform bool
}

// supervisoryControl defines the software control interface for the ESS Asset
type supervisoryControl struct {
	Enable bool
	Manual bool
}

// Config holds the ESS asset configuration parameters
type Config struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
	RatedKWH  float64 `json:"RatedKWH"`
}

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	status := Status{}
	control := Control{
		MachineControl{false, 0, 0, false},
		supervisoryControl{false, false},
	}
	return Asset{&sync.Mutex{}, PID, device, status, control, config}, err

}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.setStatus)
}

func (a *Asset) setStatus(s Status) {
	if s.Timestamp > a.status.Timestamp { // mux before?
		a.mux.Lock()
		defer a.mux.Unlock()
		a.status = s
	}
}

// WriteControl requests a physical device write of the data held in the GridAsset control field.
func (a Asset) WriteControl() {
	a.mux.Lock()
	defer a.mux.Unlock()
	control := a.control.machine
	go a.device.WriteDeviceControl(control)
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// PID is a getter for the asset PID
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Name is a getter for the asset Name
func (a Asset) Name() string {
	return a.config.Name
}

// KW returns the asset's measured real power
func (a Asset) KW() float64 {
	return a.status.KW
}

// KVAR returns the asset's measured reactive power
func (a Asset) KVAR() float64 {
	return a.status.KVAR
}

// KWCmd sets the asset's real power setpoint
func (a *Asset) KWCmd(kw float64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.machine.KW = kw
}

// KVARCmd sets the asset's reactive power setpoint
func (a *Asset) KVARCmd(kvar float64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.machine.KVAR = kvar
}

// RunCmd sets the asset's run request state
func (a *Asset) RunCmd(run bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.machine.Run = run
}

// Gridform sets the asset's gridform request state
func (a *Asset) Gridform(gridform bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.machine.Gridform = gridform
}
