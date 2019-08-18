package grid

import (
	"encoding/json"
	"io/ioutil"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/grid/sel1547"
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
	Static  StaticConfig  `json:"StaticConfig"`
	Dynamic DynamicConfig `json:"DynamicConfig"`
}

// StaticConfig is a data structure representing an architypical fixed Grid Intertie configuration
type StaticConfig struct {
	Name      string  `json:"Name"`
	Target    string  `json:"Target"`
	KwRated   float64 `json:"KwRated"`
	KvarRated float64 `json:"KvarRated"`
}

// DynamicConfig is a data structure representing an architypical adjustable Grid Intertie configuration
type DynamicConfig struct {
	DemandLimit float64 `json:"DemandLimit"`
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

func New(configPath string) (Asset, error) {
	configFileData := readConfig(configPath)

	config := Config{}
	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		return Asset{}, err
	}

	device, err := target(config)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	status := Status{}
	control := Control{}

	return Asset{PID, device, status, control, config}

}

func target(config map[string]interface{}) (asset.Device, error) {
	switch config {
	case "sel1547":
		return sel1547.New(map[string]interface{})
	}
	return nil, error.New("unrecongized target grid device: %v", config)
}

func readConfig(contfigPath string) (map[string]interface{}, error) {
	var config map[string]interface{}
	jsonFile, err := ioutil.ReadFile(contfigPath + ".json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
