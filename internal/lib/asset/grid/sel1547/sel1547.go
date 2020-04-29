package sel1547

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/comm/modbuscomm"
)

// SEL1547 target
type SEL1547 struct {
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the SEL1547
type Status struct {
	KW                   int  `json:"KW"`
	KVAR                 int  `json:"KVAR"`
	PositiveRealCapacity int  `json:"PositiveRealCapacity"`
	NegativeRealCapacity int  `json:"NegativeRealCapacity"`
	Synchronized         bool `json:"Synchronized"`
}

// Control data structure for the SEL1547
type Control struct {
	closeIntertie int
}

// Comm data structure for the SEL1547
type Comm struct {
	TargetConfig modbuscomm.PollerConfig `json:"TargetConfig"`
	handler      modbuscomm.ModbusComm
	Registers    []modbuscomm.Register `json:"Registers"`
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a SEL1547) ReadDeviceStatus() (interface{}, error) {
	response, err := a.comm.read()
	err = a.status.update(response)
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a SEL1547) WriteDeviceControl(c interface{}) error {
	a.Control(c.(grid.Control))
	payload, err := a.control.payload()
	if err != nil {
		return err
	}

	err = a.comm.write(payload)
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a SEL1547) Status() grid.Status {
	// map deviceStatus to GridStatus
	return grid.Status{
		KW:           float64(a.status.KW),
		KVAR:         float64(a.status.KVAR),
		Synchronized: bool(a.status.Synchronized),
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a SEL1547) Control(c grid.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		closeIntertie: btoi(c.CloseIntertie),
	}

	a.control = updatedControl
}

// update unmarshals a JSON response into the sel1547 status
func (a *Status) update(response []byte) error {
	updatedStatus := &Status{}
	err := json.Unmarshal(response, updatedStatus)
	if err != nil {
		return err
	}

	a = updatedStatus
	return err
}

// payload marshals a JSON string from sel1547 control
func (c Control) payload() ([]byte, error) {
	payload, err := json.Marshal(c)
	return payload, err
}

func (c Comm) read() ([]byte, error) {
	registers := modbuscomm.FilterRegisters(c.Registers, "read-only")
	response, err := c.handler.Read(registers)
	return response, err
}

func (c Comm) write(payload []byte) error {
	registers := modbuscomm.FilterRegisters(c.Registers, "write-only")
	err := c.handler.Write(registers, payload)
	return err
}

// New returns an initalized SEL1547 Asset; this is part of the Asset interface.
func New(configPath string) (grid.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return grid.Asset{}, err
	}

	commConfig, err := readCommConfig(jsonConfig)
	if err != nil {
		return grid.Asset{}, err
	}

	device := SEL1547{
		status: Status{
			KW:                   0,
			KVAR:                 0,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
			Synchronized:         false,
		},
		control: Control{
			closeIntertie: 0,
		},
		comm: commConfig,
	}

	return grid.New(jsonConfig, device)
}

func readCommConfig(config []byte) (Comm, error) {
	c := Comm{}
	err := json.Unmarshal(config, &c)
	if err != nil {
		return c, err
	}

	c.handler = modbuscomm.NewPoller(c.TargetConfig)
	return c, nil
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
