package ess

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/ess"
)

// Asset is a data structure for an ESS Asset
type Asset struct {
	pid     uuid.UUID
	device  Device
	status  Status
	control Control
	config  Config
}

type Device interface {
	ReadDeviceStatus(func(ess.Status))
	WriteDeviceControl(Control)
}

// Status is a data structure representing an architypical ESS status
type Status struct {
	mutex                *sync.Mutex
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

// Control is a data structure representing an architypical ESS control
type Control struct {
	mutex    *sync.Mutex
	Run      bool
	Enable   bool
	KW       float64
	KVAR     float64
	GridForm bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Static StaticConfig `json:"StaticConfig"`
}

// StaticConfig is a data structure representing an architypical fixed ESS configuration
type StaticConfig struct { // Should this get transfered over to the specific class?
	Name      string  `json:"Name"`
	KWRated   float64 `json:"KWRated"`
	KVARRated float64 `json:"KVARRated"`
	KWHRated  float64 `json:"KWHRated"`
}

// PID is a getter for the ess.Asset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Status is a getter for the ess.Asset status field
func (a Asset) Status() Status {
	return a.status
}

// SetControl is a setter for the ess.Asset control field
func (a *Asset) SetControl(c Control) {
	a.control.mutex.Lock()
	defer a.control.mutex.Unlock()
	a.control = c
}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() error {
	go a.device.ReadDeviceStatus(a.setStatus)
	return nil
}

func (a *Asset) setStatus(s Status) {
	a.control.mutex.Lock()
	defer a.control.mutex.Unlock()
	a.status = s
}

// WriteControl requests a physical device write of the data held in the GridAsset control field.
func (a Asset) WriteControl() error {
	go a.device.WriteDeviceControl(a.control)
	return nil
}

// New returns a configured Asset
func New(jsonConfig []byte, device Device) (Asset, error) {
	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	status := Status{
		mutex: &sync.Mutex{},
	}
	control := Control{
		mutex: &sync.Mutex{},
	}
	return Asset{PID, device, status, control, config}, err

}
