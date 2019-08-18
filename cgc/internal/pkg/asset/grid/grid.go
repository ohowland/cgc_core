package grid

import (
	"encoding/json"
	"io/ioutil"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

// Asset is a datastructure for an Energy Storage System Asset
type Asset struct {
	PID     uuid.UUID
	device  asset.Device
	status  Status
	control Control
	config  Config
}

// Status is a data structure representing an architypical Grid Intertie status
type Status struct {
	Kw   float64
	Kvar float64
}

// Control is a data structure representing an architypical Grid Intertie control
type Control struct {
	CloseIntertie bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	static  StaticConfig
	dynamic DynamicConfig
}

// StaticConfig is a data structure representing an architypical fixed Grid Intertie configuration
type StaticConfig struct {
	Name      string
	KwRated   float64
	KvarRated float64
}

// DynamicConfig is a data structure representing an architypical adjustable Grid Intertie configuration
type DynamicConfig struct {
	DemandLimit float64
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

func readConfig(path string) (Config, error) {
	c := Config{}
	jsonFile, err := ioutil.ReadFile(path + ".json")
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(jsonFile, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}
