package virtualpv

import (
	"io/ioutil"
	"log"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/pv"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

// VirtualPV target
type VirtualPV struct {
	pid       uuid.UUID
	status    Status
	control   Control
	comm      Comm
	observers Observers
}

// Status data structure for the VirtualPV
type Status struct {
	KW     float64 `json:"KW"`
	KVAR   float64 `json:"KVAR"`
	Hz     float64 `json:"Hz"`
	Volt   float64 `json:"Volt"`
	Online bool    `json:"Online"`
}

// Control data structure for the VirtualPV
type Control struct {
	run bool
}

type Config struct {
}

// Comm data structure for the VirtualPV
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

type Observers struct {
	assetObserver chan<- virtualacbus.Source
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualPV) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualPV) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps grid.DeviceStatus to grid.Status
func (a VirtualPV) Status() pv.Status {
	// map deviceStatus to GridStatus
	return pv.Status{
		KW:     a.status.KW,
		KVAR:   a.status.KVAR,
		Volt:   a.status.Volt,
		Hz:     a.status.Hz,
		Online: a.status.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func (a VirtualPV) Control(c pv.Control) {
	// map GridControl params to deviceControl

	updatedControl := Control{
		run: c.RunRequest,
	}

	a.control = updatedControl
}

func (a *VirtualPV) read() error {
	select {
	case in := <-a.comm.incoming:
		a.status = in
		//log.Printf("[VirtualESS: read status: %v]", in)
	default:
		log.Println("[VirtualPV: read failed]")
	}
	return nil
}

func (a *VirtualPV) write() error {
	select {
	case a.comm.outgoing <- a.control:
		//log.Printf("[VirtualESS: write control: %v]", a.control)
	default:
		log.Println("[VirtualPV: write failed]")

	}
	return nil
}

// New returns an initalized VirtualPV Asset; this is part of the Asset interface.
func New(configPath string, bus *virtualacbus.VirtualACBus) (pv.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ess.Asset{}, err
	}

	// TODO: Troubleshoot why this cannot be set to 0 length
	in := make(chan Status, 1)
	out := make(chan Control, 1)

	pid, err := uuid.NewUUID()

	device := VirtualPV{
		pid: pid,
		status: Status{
			KW:     0,
			KVAR:   0,
			Hz:     0,
			Volt:   0,
			Online: false,
		},
		control: Control{
			run: false,
		},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
		observers: Observers{
			assetObserver: bus.AssetObserver(),
		},
	}

	go virtualDeviceLoop(device.comm, device.observers)
	return pv.New(jsonConfig, device)
}

func virtualDeviceLoop(comm Comm, obs Observers) {
	dev := &VirtualPV{} // The virtual 'hardware' device
	//sm := &stateMachine{offState{}}
	var ok bool
	for {
		select {
		case dev.control, ok = <-comm.outgoing: // write to 'hardware'
			if !ok {
				break
			}
		case comm.incoming <- dev.status: // read from 'hardware'
			dev.updateObservers(obs)
			//dev.status = sm.run(*dev)
			//log.Printf("[VirtualESS-Device: state: %v]\n",
			//reflect.TypeOf(sm.currentState).String())
		}
		log.Println("[VirtualESS-Device: shutdown]")
	}
}

func (a *VirtualPV) updateObservers(obs Observers) {
	obs.assetObserver <- a.asSource()
}

func (a VirtualPV) asSource() virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: false,
	}
}
