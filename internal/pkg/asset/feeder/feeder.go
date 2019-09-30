package feeder

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

// Asset is a data structure for an Feeder Asset
type Asset struct {
	pid     uuid.UUID
	device  asset.Device
	status  Status
	control Control
	config  Config
}

// Status is a data structure representing an architypical Feeder status
type Status struct {
	KW     float64
	KVAR   float64
	Hz     float64
	Volt   float64
	Online bool
}

// Control is a data structure representing an architypical Feeder control
type Control struct {
	CloseFeeder bool
	Enable      bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Static  StaticConfig  `json:"StaticConfig"`
	Dynamic DynamicConfig `json:"DynamicConfig"`
}

// StaticConfig is a data structure representing an architypical fixed ESS configuration
type StaticConfig struct { // Should this get transfered over to the specific class?
	Name      string  `json:"Name"`
	KWRated   float64 `json:"KWRated"`
	KVARRated float64 `json:"KVARRated"`
}

// DynamicConfig is a data structure representing an architypical adjustable ESS configuration
type DynamicConfig struct{}

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
	a.control = c
}

// Control is a getter for the ess.Asset control field
func (a Asset) Control() Control {
	return a.control
}

// SetDynamicConfig is a setter for the ess.Asset dynamic config field
func (a *Asset) SetDynamicConfig(c DynamicConfig) {
	a.config.Dynamic = c
}

// StaticConfig is a getter for the ess.Asset static config field
func (a Asset) StaticConfig() StaticConfig {
	return a.config.Static
}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() error {
	status, err := a.device.ReadDeviceStatus()
	if err != nil {
		return err
	}

	a.status = status.(Status)
	return err
}

// WriteControl requests a physical device write of the data held in the GridAsset control field.
func (a Asset) WriteControl() error {
	err := a.device.WriteDeviceControl(a.control)
	return err
}

// New returns a configured Asset
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
