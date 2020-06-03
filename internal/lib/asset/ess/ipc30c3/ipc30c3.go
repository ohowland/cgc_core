package ipc30c3

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc_core/internal/pkg/asset/ess"
	"github.com/ohowland/cgc_core/internal/pkg/comm/modbuscomm"
)

// IPC30C3 target
type IPC30C3 struct {
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the IPC30C3
type Status struct {
	KW                   int `json:"KW"`
	KVAR                 int `json:"KVAR"`
	SOC                  int `json:"SOC"`
	PositiveRealCapacity int `json:"PositiveRealCapacity"`
	NegativeRealCapacity int `json:"NegativeRealCapacity"`
}

// Control data structure for the IPC30C3
type Control struct {
	runRequest int
}

// Comm data structure for the IPC30C3
type Comm struct {
	TargetConfig modbuscomm.PollerConfig `json:"TargetConfig"`
	handler      modbuscomm.ModbusComm
	Registers    []modbuscomm.Register `json:"Registers"`
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a IPC30C3) ReadDeviceStatus() (interface{}, error) {
	response, err := a.comm.read()
	err = a.status.update(response)
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a IPC30C3) WriteDeviceControl(c interface{}) error {
	a.Control(c.(ess.Control))
	payload, err := a.control.payload()
	if err != nil {
		return err
	}

	err = a.comm.write(payload)
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a IPC30C3) Status() ess.Status {
	// map deviceStatus to GridStatus
	return ess.Status{
		KW:                   float64(a.status.KW),
		KVAR:                 float64(a.status.KVAR),
		SOC:                  float64(a.status.SOC),
		PositiveRealCapacity: float64(a.status.PositiveRealCapacity),
		NegativeRealCapacity: float64(a.status.NegativeRealCapacity),
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a IPC30C3) Control(c ess.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		runRequest: btoi(c.RunRequest),
	}

	a.control = updatedControl
}

// update unmarshals a JSON response into the IPC30C3 status
func (a *Status) update(response []byte) error {
	updatedStatus := &Status{}
	err := json.Unmarshal(response, updatedStatus)
	if err != nil {
		return err
	}

	a = updatedStatus
	return err
}

// payload marshals a JSON string from IPC30C3 control
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

// New returns an initalized IPC30C3 Asset; this is part of the Asset interface.
func New(configPath string) (ess.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	commConfig, err := readCommConfig(jsonConfig)
	if err != nil {
		return ess.Asset{}, err
	}

	device := IPC30C3{
		status: Status{
			KW:                   0,
			KVAR:                 0,
			SOC:                  0,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
		},
		control: Control{
			runRequest: 0,
		},
		comm: commConfig,
	}

	return ess.New(jsonConfig, device)
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
