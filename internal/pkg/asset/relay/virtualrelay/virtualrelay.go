package virtualrelay

import (
	"io/ioutil"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/ohowland/cgc/internal/pkg/asset/relay"
)

const (
	queueSize = 50
)

// VirtualRelay target
type VirtualRelay struct {
	id      uuid.UUID
	status  Status
	control Control
	comm    Comm
}

// Status data structure for the VirtualRelay
type Status struct {
	Hz   float64 `json:"Hz"`
	Volt float64 `json:"Volt"`
}

// Control data structure for the VirtualRelay
type Control struct{}

type Config struct{}

// Comm data structure for the VirtualRelay
type Comm struct {
	incoming chan Status
	outgoing chan Control
}

// ReadDeviceStatus requests a physical device read over the communication interface
func (a VirtualRelay) ReadDeviceStatus() (interface{}, error) {
	err := a.read()
	return a.Status(), err
}

// WriteDeviceControl prequests a physical device write over the communication interface
func (a VirtualRelay) WriteDeviceControl(c interface{}) error {
	err := a.write()
	return err
}

// Status maps relay.DeviceStatus to relay.Status
func (a VirtualRelay) Status() relay.Status {
	return relay.Status{
		Hz:   float64(a.status.Hz),
		Volt: float64(a.status.Volt),
	}
}

// Control maps relay.Control to bus.DeviceControl
func (a VirtualRelay) Control(c relay.Control) {

	updatedControl := Control{}

	a.control = updatedControl
}

func (a *VirtualRelay) read() error {
	select {
	case in := <-a.comm.incoming:
		a.status = in
		//log.Printf("[VirtualRelay: read status: %v]", in)
	default:
		log.Println("[VirtualRelay: read failed]")
	}
	return nil
}

func (a *VirtualRelay) write() error {
	select {
	case a.comm.outgoing <- a.control:
		//log.Printf("[VirtualRelay: write control: %v]", a.control)
	default:
		log.Println("[VirtualRelay: write failed]")

	}
	return nil
}

// New returns an initalized VirtualRelay Asset; this is part of the Asset interface.
func New(configPath string) (relay.Asset, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return relay.Asset{}, err
	}

	in := make(chan Status, 1)
	out := make(chan Control, 1)

	sources := make(chan Source, queueSize)
	gridformer := make(chan Source, 1)

	id, err := uuid.NewUUID()

	device := VirtualRelay{
		id: id,
		status: Status{
			Hz:    0.0,
			Volts: 0.0,
		},
		control: Control{},
		comm: Comm{
			incoming: in,
			outgoing: out,
		},
	}

	go virtualDeviceLoop(device.comm)
	return relay.New(jsonConfig, device)
}

func virtualDeviceLoop(comm Comm) {
	dev := &VirtualRelay{} // The virtual 'hardware' device
	sm := &stateMachine{offState{}}
	var ok bool
	for {
		select {
		case dev.control, ok = <-comm.outgoing: // write to virtual device
			if !ok {
				break
			}
		case comm.incoming <- dev.status: // read from virtual device
		default:
			updateVirtualDevice(dev, comm)
		}
	}
	log.Println("[VirtualRelay-Device: shutdown]")
}

func updateVirtualDevice(dev *VirtualRelay, comm Comm) {
	dev = updateVirtualDevice(dev, comm)
	log.Printf("[VirtualRelay-Device: state: %v]\n",
		reflect.TypeOf(sm.currentState).String())
	time.Sleep(time.Duration(200) * time.Millisecond)
}
