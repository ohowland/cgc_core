package asset

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/comm"
)

// SEL1547 target
type SEL1547 struct {
	deviceStatus       sel1547Status
	deviceControl      sel1547Control
	deviceStaticConfig sel1547StaticConfig
	deviceComm         sel1547Comm
}

// Status fulfills the Grid Status Interface
func (a SEL1547) Status() (asset.GridStatus, error) {
	// map deviceStatus to GridStatus
	return asset.GridStatus{
		Kw:   float64(a.deviceStatus.Kw),
		Kvar: float64(a.deviceStatus.Kvar),
	}, nil
}

// Control fulfills the Grid Control Interface
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

// Config fulfills the Grid Static Config Interface
func (a SEL1547) Config() (asset.GridStaticConfig, error) {
	// map deviceConfig to GridConfig

	return asset.GridStaticConfig{
		Name:      a.deviceStaticConfig.Name,
		KwRated:   float64(a.deviceStaticConfig.KwRated),
		KvarRated: float64(a.deviceStaticConfig.KvarRated),
	}, nil
}

// ReadDeviceStatus provides the API to read status from a device through the Asset's comm interface
func (a SEL1547) ReadDeviceStatus() error {
	response, err := a.deviceComm.read()
	err = a.deviceStatus.update(response)
	return err
}

// WriteDeviceControl provies the API to write control to a device through the Asset's comm interface
func (a SEL1547) WriteDeviceControl() error {
	payload, err := a.deviceControl.payload()
	if err != nil {
		return err
	}
	err = a.deviceComm.write(payload)
	return err
}

type sel1547Status struct {
	Kw           int  `json:"Kw"`
	Kvar         int  `json:"Kvar"`
	Synchronized bool `json:"Synchronized"`
}

// update unmarshals a JSON response into the sel1547 status
func (a sel1547Status) update(response []byte) error {
	err := json.Unmarshal(response, &a)
	return err
}

type sel1547Control struct {
	closeIntertie int
}

// payload marshals a JSON string from sel1547control
func (c sel1547Control) payload() ([]byte, error) {
	payload, err := json.Marshal(c)
	return payload, err
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
	TargetConfig comm.PollerConfig `json:"Config"`
	handler      comm.ModbusComm
	Registers    []comm.Register `json:"Registers"`
}

func (c sel1547Comm) read() ([]byte, error) {
	registers := comm.FilterRegisters(c.Registers, "read-only")
	response, err := c.handler.Read(registers)
	return response, err
}

func (c sel1547Comm) write(payload []byte) error {
	registers := comm.FilterRegisters(c.Registers, "write-only")
	err := c.handler.Write(registers, payload)
	return err
}

// NewAsset returns an initalized SEL1547 Asset
func NewAsset(configPath string) asset.GridAsset {
	staticConfig, err := readStaticConfig(configPath + "_static")
	if err != nil {
		panic(err)
	}

	commConfig, err := readCommConfig(configPath + "_comm")
	if err != nil {
		panic(err)
	}

	return SEL1547{
		deviceStatus: sel1547Status{
			Kw:           0,
			Kvar:         0,
			Synchronized: false,
		},
		deviceControl: sel1547Control{
			closeIntertie: 0,
		},
		deviceStaticConfig: staticConfig,
		deviceComm:         commConfig,
	}
}

func readStaticConfig(path string) (sel1547StaticConfig, error) {
	c := sel1547StaticConfig{}
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(jsonFile, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func readCommConfig(path string) (sel1547Comm, error) {
	c := sel1547Comm{}
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(jsonFile, &c)
	if err != nil {
		return c, err
	}

	c.handler = comm.NewPoller(c.TargetConfig)
	return c, nil
}
