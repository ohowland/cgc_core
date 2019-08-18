package sel1547

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/comm"
)

// SEL1547 target
type SEL1547 struct {
	status  Status
	control Control
	comm    Comm
}

type Status struct {
	Kw           int  `json:"Kw"`
	Kvar         int  `json:"Kvar"`
	Synchronized bool `json:"Synchronized"`
}

type Control struct {
	closeIntertie int
}

type Comm struct {
	TargetConfig comm.PollerConfig `json:"Config"`
	handler      comm.ModbusComm
	Registers    []comm.Register `json:"Registers"`
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a SEL1547) ReadDeviceStatus() (interface{}, error) {
	response, err := a.comm.read()
	err = a.status.update(response)
	return a.status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a SEL1547) WriteDeviceControl(c interface{}) error {
	a.control(c.(grid.Control))
	payload, err := a.control.payload()
	if err != nil {
		return err
	}
	err = a.comm.write(payload)
	return err
}

// Status fulfills the Grid Status Interface
func (a SEL1547) status() grid.Status {
	// map deviceStatus to GridStatus
	return grid.Status{
		Kw:   float64(a.deviceStatus.Kw),
		Kvar: float64(a.deviceStatus.Kvar),
	}
}

// Control fulfills the Grid Control Interface
func (a SEL1547) control(c grid.Control) {
	// map GridControl params to deviceControl
	mapCloseIntertie := 0
	if c.CloseIntertie != false {
		mapCloseIntertie = 1
	}

	updatedControl := sel1547Control{
		closeIntertie: mapCloseIntertie,
	}

	a.deviceControl = updatedControl
}

// update unmarshals a JSON response into the sel1547 status
func (c Status) update(response []byte) error {
	err := json.Unmarshal(response, &a)
	return err
}

// payload marshals a JSON string from sel1547 control
func (c Control) payload() ([]byte, error) {
	payload, err := json.Marshal(c)
	return payload, err
}

func (c Comm) read() ([]byte, error) {
	registers := comm.FilterRegisters(c.Registers, "read-only")
	response, err := c.handler.Read(registers)
	return response, err
}

func (c Comm) write(payload []byte) error {
	registers := comm.FilterRegisters(c.Registers, "write-only")
	err := c.handler.Write(registers, payload)
	return err
}

// New returns an initalized SEL1547 Asset; this is part of the Asset interface.
func New(configPath string) (*SEL1547, error) {
	commConfig, err := readCommConfig(configPath + "_comm")
	if err != nil {
		return &SEL1547{}, err
	}

	return &SEL1547{
		deviceStatus: sel1547Status{
			Kw:           0,
			Kvar:         0,
			Synchronized: false,
		},
		deviceControl: sel1547Control{
			closeIntertie: 0,
		},
		deviceComm: commConfig,
	}, nil
}

func readCommConfig(path string) (Comm, error) {
	c := Comm{}
	jsonFile, err := ioutil.ReadFile(path + ".json")
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
