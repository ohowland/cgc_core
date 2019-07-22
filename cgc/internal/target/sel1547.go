package target

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc/internal/asset"
	"github.com/ohowland/cgc/internal/comm"
)

// SEL1547 target
type SEL1547 struct {
	deviceStatus       sel1547Status
	deviceControl      sel1547Control
	deviceStaticConfig sel1547StaticConfig
	deviceComm         sel1547Comm
}

type sel1547Status struct {
	kw           int
	kvar         int
	synchronized bool
}

type sel1547Control struct {
	closeIntertie int
}

type sel1547StaticConfig struct {
	Name      string `json:"Name"`
	KwRated   int    `json:"KwRated"`
	KvarRated int    `json:"KvarRated"`
}

type sel1547DynamicCondig struct {
	pid uint16
}

type sel1547Comm struct {
	handler   comm.ModbusComm
	registers comm.Register
}

func (a SEL1547) Status() (asset.GridStatus, error) {
	// map deviceStatus to GridStatus
	return asset.GridStatus{
		Kw:   float64(a.deviceStatus.kw),
		Kvar: float64(a.deviceStatus.kvar),
	}, nil
}

func (a SEL1547) Control(ctrl asset.GridControl) error {
	// map GridControl params to deviceControl
	mapCloseIntertie := 0
	if ctrl.CloseIntertie != false {
		mapCloseIntertie = 1
	}

	updatedControl := sel1547Control{
		closeIntertie: mapCloseIntertie,
	}

	a.deviceControl = updatedControl
	return nil
}

func (a SEL1547) Config() (asset.GridStaticConfig, error) {
	// map deviceConfig to GridConfig

	return asset.GridStaticConfig{
		Name:      a.deviceStaticConfig.Name,
		KwRated:   float64(a.deviceStaticConfig.KwRated),
		KvarRated: float64(a.deviceStaticConfig.KvarRated),
	}, nil
}

// NewAsset returns an initalized SEL1547 Asset
func NewAsset(configPath string) (asset.GridAsset, error) {
	staticConfig := sel1547StaticConfig{}
	staticConfig.readConfig(configPath)

	return SEL1547{
		deviceStatus: sel1547Status{
			kw:           0,
			kvar:         0,
			synchronized: false,
		},
		deviceControl: sel1547Control{
			closeIntertie: 0,
		},
		deviceStaticConfig: staticConfig,
	}, nil
}

func (c *sel1547StaticConfig) readConfig(path string) *sel1547StaticConfig {
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(jsonFile, &c)
	if err != nil {
		return nil
	}
	return c
}
