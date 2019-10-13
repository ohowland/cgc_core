package grid

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

// Asset is a datastructure for an Energy Storage System Asset
type Asset struct {
	mux     sync.Mutex
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

// Status is a data structure representing an architypical Grid Intertie status
type Status struct {
	Timestamp            int64
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volts                float64
	PositiveRealCapacity float64
	NegativeRealCapacity float64
	Synchronized         bool
	Online               bool
}

// Control is a data structure representing an architypical Grid Intertie control
type Control struct {
	dispatch    MachineControl
	operator    MachineControl
	supervisory SupervisoryControl
}

type MachineControl struct {
	mux           sync.Mutex
	CloseIntertie bool
}

type SupervisoryControl struct {
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
	return Asset{sync.Mutex{}, PID, device, status, control, config}, err

}

// UpdateStatus requests a physical device read and updates the GridAsset status field
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
	var control MachineControl
	if a.control.supervisory.Manual {
		a.control.operator.mux.Lock()
		defer a.control.operator.mux.Unlock()
		control = a.control.operator
	} else {
		a.control.dispatch.mux.Lock()
		defer a.control.dispatch.mux.Unlock()
		control = a.control.dispatch
	}
	go a.device.WriteDeviceControl(control)
}

// PID is a getter for the GridAsset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

func (a Asset) Name() string {
	return a.config.Name
}

func (a Asset) KW() float64 {
	return a.status.KW
}

func (a Asset) KVAR() float64 {
	return a.status.KVAR
}

func (a *Asset) DispatchControlHandle() asset.MachineControl {
	return &a.control.dispatch
}

func (a *Asset) OperatorControlHandle() asset.MachineControl {
	return &a.control.operator
}

// KWCmd sets the asset's real power setpoint
func (a *MachineControl) KWCmd(kw float64) {}

// KVARCmd sets the asset's reactive power setpoint
func (a *MachineControl) KVARCmd(kvar float64) {}

// RunCmd sets the asset's run request state
func (a *MachineControl) RunCmd(run bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.CloseIntertie = run
}

// GridformCmd sets the asset's gridform request state
func (a *MachineControl) GridformCmd(gridform bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.CloseIntertie = gridform
}
