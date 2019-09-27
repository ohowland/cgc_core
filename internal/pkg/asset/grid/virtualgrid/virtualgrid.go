package virtualgrid

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/comm/modbuscomm"
)

// VirtualGrid target
type VirtualGrid struct {
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualGrid
type Status struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	PositiveRealCapacity float64 `json:"PositiveRealCapacity"`
	NegativeRealCapacity float64 `json:"NegativeRealCapacity"`
	Synchronized         bool    `json:"Synchronized"`
}

// Control data structure for the VirtualGrid
type Control struct {
	closeIntertie int
}

// Comm data structure for the VirtualGrid
type Comm struct {
	TargetConfig modbuscomm.PollerConfig `json:"TargetConfig"`
	handler      modbuscomm.ModbusComm
	Registers    []modbuscomm.Register `json:"Registers"`
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualGrid) ReadDeviceStatus() (interface{}, error) {
	response, err := a.comm.read()
	err = a.status.update(response)
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualGrid) WriteDeviceControl(c interface{}) error {
	a.Control(c.(grid.Control))
	payload, err := a.control.payload()
	if err != nil {
		return err
	}

	err = a.comm.write(payload)
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a VirtualGrid) Status() grid.Status {
	// map deviceStatus to GridStatus
	return grid.Status{
		KW:           float64(a.status.KW),
		KVAR:         float64(a.status.KVAR),
		Synchronized: bool(a.status.Synchronized),
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a VirtualGrid) Control(c grid.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		closeIntertie: btoi(c.CloseIntertie),
	}

	a.control = updatedControl
}

// update unmarshals a JSON response into the VirtualGrid status
func (a *Status) update(response []byte) error {
	updatedStatus := &Status{}
	err := json.Unmarshal(response, updatedStatus)
	if err != nil {
		return err
	}

	a = updatedStatus
	return err
}

// payload marshals a JSON string from VirtualGrid control
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

	device := VirtualGrid{
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

// GridFormer implements the virtual.Power interface
func (a VirtualGrid) GridFormer() bool {
	return a.status.GridForming
}

// KW implements the asset.Power interface
func (a VirtualGrid) KW() float64 {
	return a.status.KW
}

// KVAR implements the asset.Power interface
func (a VirtualGrid) KVAR() float64 {
	return a.status.KVAR
}
