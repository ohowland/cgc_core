package grid

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

// Asset is a datastructure for an Energy Storage System Asset
type Asset struct {
	pid     uuid.UUID
	device  asset.Device
	status  Status
	control Control
	config  Config
}

// Status is a data structure representing an architypical Grid Intertie status
type Status struct {
	KW                   float64
	KVAR                 float64
	PostiveRealCapacity  float64
	NegativeRealCapacity float64
	Synchronized         bool
}

// Control is a data structure representing an architypical Grid Intertie control
type Control struct {
	CloseIntertie bool
	Enable        bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Static  StaticConfig  `json:"StaticConfig"`
	Dynamic DynamicConfig `json:"DynamicConfig"`
}

// StaticConfig is a data structure representing an architypical fixed Grid Intertie configuration
type StaticConfig struct {
	Name      string  `json:"Name"`
	KWRated   float64 `json:"KwRated"`
	KVARRated float64 `json:"KvarRated"`
}

// DynamicConfig is a data structure representing an architypical adjustable Grid Intertie configuration
type DynamicConfig struct {
	DemandLimit float64 `json:"DemandLimit"`
}

// PID is a getter for the GridAsset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Status is a getter for the GridAsset status field
func (a Asset) Status() Status {
	return a.status
}

// Control is a setter for the GridAsset control field
func (a *Asset) Control(c Control) {
	a.control = c
}

// Config is a setter for the GridAsset config field
func (a *Asset) Config(c Config) {
	a.config = c
}

// UpdateStatus requests a physical device read and updates the GridAsset status field
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