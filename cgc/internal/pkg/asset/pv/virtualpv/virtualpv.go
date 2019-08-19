package virtualpv

import (
	"encoding/json"
	"io/ioutil"

	"github.com/ohowland/cgc/internal/pkg/asset/pv"
	"github.com/ohowland/cgc/internal/pkg/comm"
)

// VirtualPV target
type VirtualPV struct {
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualPV
type Status struct {
	KW   int `json:"KW"`
	KVAR int `json:"KVAR"`
}

// Control data structure for the VirtualPV
type Control struct {
	RunRequest int
}

// Comm data structure for the VirtualPV
type Comm struct {
	TargetConfig comm.PollerConfig `json:"TargetConfig"`
	handler      comm.ModbusComm
	Registers    []comm.Register `json:"Registers"`
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualPV) ReadDeviceStatus() (interface{}, error) {
	response, err := a.comm.read()
	err = a.status.update(response)
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualPV) WriteDeviceControl(c interface{}) error {
	a.Control(c.(pv.Control))
	payload, err := a.control.payload()
	if err != nil {
		return err
	}

	err = a.comm.write(payload)
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a VirtualPV) Status() pv.Status {
	// map deviceStatus to GridStatus
	return pv.Status{
		KW:   float64(a.status.KW),
		KVAR: float64(a.status.KVAR),
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a VirtualPV) Control(c pv.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		RunRequest: btoi(c.RunRequest),
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
	registers := comm.FilterRegisters(c.Registers, "read-only")
	response, err := c.handler.Read(registers)
	return response, err
}

func (c Comm) write(payload []byte) error {
	registers := comm.FilterRegisters(c.Registers, "write-only")
	err := c.handler.Write(registers, payload)
	return err
}

// New returns an initalized VirtualPV Asset; this is part of the Asset interface.
func New(configPath string) (pv.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return pv.Asset{}, err
	}

	commConfig, err := readCommConfig(jsonConfig)
	if err != nil {
		return pv.Asset{}, err
	}

	device := VirtualPV{
		status: Status{
			KW:   0,
			KVAR: 0,
		},
		control: Control{
			RunRequest: 0,
		},
		comm: commConfig,
	}

	return pv.New(jsonConfig, device)
}

func readCommConfig(config []byte) (Comm, error) {
	c := Comm{}
	err := json.Unmarshal(config, &c)
	if err != nil {
		return c, err
	}

	c.handler = comm.NewPoller(c.TargetConfig)
	return c, nil
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
