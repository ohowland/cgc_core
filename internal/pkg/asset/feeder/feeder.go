package feeder

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for an Feeder Asset
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

// Status is a data structure representing an architypical Feeder status
type Status struct {
	Timestamp int64
	KW        float64
	KVAR      float64
	Hz        float64
	Volt      float64
	Online    bool
}

// Control is a data structure representing an architypical Feeder control
type Control struct {
	machine     MachineControl
	supervisory supervisoryControl
}

type MachineControl struct {
	CloseFeeder bool
}

type supervisoryControl struct {
	Enable bool
	Manual bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
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
	control := Control{}

	return Asset{&sync.Mutex{}, PID, device, status, control, config}, err
}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.setStatus)
}

func (a *Asset) setStatus(s Status) {
	if s.Timestamp > a.status.Timestamp {
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

func (a Asset) DeviceController() DeviceController {
	return a.device
}

// PID is a getter for the ess.Asset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Name of asset
func (a Asset) Name() string {
	return a.config.Name
}

// KW reported by asset
func (a Asset) KW() float64 {
	return a.status.KW
}

// KVAR reported by asset
func (a Asset) KVAR() float64 {
	return a.status.KVAR
}

// KWCmd sets the asset's real power setpoint
func (a *Asset) KWCmd(kw float64) {}

// KVARCmd sets the asset's reactive power setpoint
func (a *Asset) KVARCmd(kvar float64) {}

// RunCmd sets the asset's run request state
func (a *Asset) RunCmd(run bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.machine.CloseFeeder = run
}
