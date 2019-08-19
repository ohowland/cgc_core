package virtualess

import (
	"io/ioutil"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
)

// VirtualESS target
type VirtualESS struct {
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualESS
type Status struct {
	KW                   int `json:"KW"`
	KVAR                 int `json:"KVAR"`
	SOC                  int `json:"SOC"`
	PositiveRealCapacity int `json:"PositiveRealCapacity"`
	NegativeRealCapacity int `json:"NegativeRealCapacity"`
}

// Control data structure for the VirtualESS
type Control struct {
	runRequest bool
}

// Comm data structure for the VirtualESS
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualESS) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualESS) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a VirtualESS) Status() ess.Status {
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
func (a VirtualESS) Control(c ess.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		runRequest: c.RunRequest,
	}

	a.control = updatedControl
}

func (a *VirtualESS) read() error {
	in := <-a.comm.incoming
	a.status = in
	return nil
}

func (a *VirtualESS) write() error {
	a.comm.outgoing <- a.control
	return nil
}

// New returns an initalized IPC30C3 Asset; this is part of the Asset interface.
func New(configPath string) (ess.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	in := make(chan Status)
	out := make(chan Control)

	device := VirtualESS{
		status: Status{
			KW:                   0,
			KVAR:                 0,
			SOC:                  0,
			PositiveRealCapacity: 0,
			NegativeRealCapacity: 0,
		},
		control: Control{
			runRequest: false,
		},
		comm: Comm{in, out},
	}

	ess, err := ess.New(jsonConfig, device)

	go launchVirtualDevice(device, ess.Config())
	return ess, err
}

func launchVirtualDevice(d VirtualESS, cfg ess.Config) {

	for {
		select {
		case in := <-d.comm.outgoing:
			d.control = in
			d.status = statemachine(d)
			d.comm.incoming <- d.status
		default:
		}
	}
}

func statemachine(d VirtaulESS, cfg ess.Config) {

}
