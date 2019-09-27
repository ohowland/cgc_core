package pv

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

// Asset is a datastructure for an PV Asset
type Asset struct {
	pid     uuid.UUID
	device  asset.Device
	status  Status
	control Control
	config  Config
}

// Status is a data structure representing an architypical PV status
type Status struct {
	KW   float64
	KVAR float64
}

// Control is a data structure representing an architypical PV control
type Control struct {
	RunRequest bool
	Enable     bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Static  StaticConfig  `json:"StaticConfig"`
	Dynamic DynamicConfig `json:"DynamicConfig"`
}

// StaticConfig is a data structure representing an architypical fixed PV configuration
type StaticConfig struct {
	Name      string  `json:"Name"`
	KWRated   float64 `json:"KwRated"`
	KVARRated float64 `json:"KvarRated"`
}

// DynamicConfig is a data structure representing an architypical adjustable PV configuration
type DynamicConfig struct {
}

// PID is a getter for the unique identifier field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Status is a getter for the pv.Asset status field
func (a Asset) Status() Status {
	return a.status
}

// Control is a setter for the pv.Asset control field
func (a *Asset) Control(c Control) {
	a.control = c
}

// Config is a setter for the pv.Asset config field
func (a *Asset) Config(c Config) {
	a.config = c
}

// UpdateStatus requests a physical device read and updates the pv.Asset status field
func (a *Asset) UpdateStatus() error {
	status, err := a.device.ReadDeviceStatus()
	if err != nil {
		return err
	}

	a.status = status.(Status)
	return err
}

// WriteControl requests a physical device write of the data held in the pv.Asset control field.
func (a Asset) WriteControl() error {
	err := a.device.WriteDeviceControl(a.control)
	return err
}

// New returns a configured pv.Asset
func New(jsonConfig []byte, device asset.Device) (Asset, error) {
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

	return Asset{PID, device, status, control, config}, err

}

// KW implements the asset.Power interface
func (a Asset) KW() float64 {
	return a.status.KW
}

// KVAR implements the asset.Power interface
func (a Asset) KVAR() float64 {
	return a.status.KVAR
}
