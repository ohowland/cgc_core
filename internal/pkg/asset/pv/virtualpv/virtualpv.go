package virtualpv

import (
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
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
	timestamp int64
	KW        float64 `json:"KW"`
	KVAR      float64 `json:"KVAR"`
	Hz        float64 `json:"Hz"`
	Volt      float64 `json:"Volt"`
	Online    bool    `json:"Online"`
}

// Control data structure for the VirtualPV
type Control struct {
	run bool
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
func (a VirtualPV) ReadDeviceStatus(setParentStatus func(pv.Status)) {
	a.status = a.read()
	setParentStatus(mapStatus(a.status)) // callback for to write parent status
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualPV) WriteDeviceControl(c pv.Control) {
	a.control = mapControl(c)
	a.write()
}

func (a VirtualPV) read() Status {
	timestamp := time.Now().UnixNano()
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	readStatus := <-a.comm.incoming
	readStatus.timestamp = timestamp
	return readStatus
}

func (a VirtualPV) write() {
	fuzzing := rand.Intn(2000)
	time.Sleep(time.Duration(fuzzing) * time.Millisecond)
	a.comm.outgoing <- a.control
}

func (a *VirtualPV) updateObservers(obs Observers) {
	obs.assetObserver <- mapSource(*a)
}

// New returns an initalized VirtualPV Asset; this is part of the Asset interface.
func New(configPath string, bus virtualacbus.VirtualACBus) (pv.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return pv.Asset{}, err
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

func mapSource(a VirtualPV) virtualacbus.Source {
	return virtualacbus.Source{
		PID:         a.pid,
		Hz:          a.status.Hz,
		Volt:        a.status.Volt,
		KW:          a.status.KW,
		KVAR:        a.status.KVAR,
		Gridforming: false,
	}
}

// Status maps grid.DeviceStatus to grid.Status
func mapStatus(s Status) pv.Status {
	// map deviceStatus to GridStatus
	return pv.Status{
		Timestamp: s.timestamp,
		KW:        s.KW,
		KVAR:      s.KVAR,
		Volt:      s.Volt,
		Hz:        s.Hz,
		Online:    s.Online,
	}
}

// Control maps grid.Control to grid.DeviceControl
func mapControl(c pv.Control) Control {
	// map GridControl params to deviceControl
	return Control{
		run: c.RunRequest,
	}
}

func virtualDeviceLoop(comm Comm, obs Observers) {
	dev := &VirtualPV{} // The virtual 'hardware' device
	//sm := &stateMachine{offState{}}
	var ok bool
loop:
	for {
		select {
		case dev.control, ok = <-comm.outgoing: // write to 'hardware'
			if !ok {
				break loop
			}
		case comm.incoming <- dev.status: // read from 'hardware'
			dev.updateObservers(obs)
			//dev.status = sm.run(*dev)
			//log.Printf("[VirtualPV-Device: state: %v]\n",
			//reflect.TypeOf(sm.currentState).String())
		}
	}
	log.Println("[VirtualPV-Device] shutdown")

}
