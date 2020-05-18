package mockasset

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// randDummyStatus returns a closure for random DummyAsset Status
func randMockStatus() func() Status {
	return func() Status {
		rand.Seed(time.Hour.Nanoseconds())
		status := Status{MachineStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}}
		return status
	}
}

func randMockControl() func() Control {
	rand.Seed(time.Hour.Nanoseconds())
	control := Control{rand.Float64(), rand.Float64()}
	return func() Control {
		return control
	}
}

func randMockConfig() func() Config {
	rand.Seed(time.Hour.Nanoseconds())
	name := fmt.Sprintf("MockAsset-%d", rand.Int())
	config := Config{name, "MockBus"}
	return func() Config {
		return config
	}
}

var AssertedStatus = randMockStatus()
var AssertedControl = randMockControl()
var AssertedConfig = randMockConfig()

type Asset struct {
	pid          uuid.UUID
	publisher    *msg.PubSub
	controlOwner uuid.UUID
	Control      Control
}

func (d Asset) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	ch, err := d.publisher.Subscribe(pid, topic)
	return ch, err
}

// Unsubscribe pid from all topic broadcasts
func (d Asset) Unsubscribe(pid uuid.UUID) {
	d.publisher.Unsubscribe(pid)
}

func (d *Asset) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	d.controlOwner = pid
	go d.controlHandler(ch)
	return nil
}

// Shutdown asset processes and cleanup resources
func (d *Asset) Shutdown(*sync.WaitGroup) error {
	return nil
}

func (d *Asset) controlHandler(ch <-chan msg.Msg) {
loop:
	for {
		data, ok := <-ch
		if !ok {
			log.Println("MockAsset controlHandler() stopping")
			break loop
		}
		control, ok := data.Payload().(Control)
		if !ok {
			log.Println("MockAsset controlHandler() bad type assertion")
			break loop
		}
		d.Control = control
	}
}

func (d Asset) PID() uuid.UUID {
	return d.pid
}

func (d Asset) Name() string {
	return fmt.Sprintf("MockAsset-%d", rand.Int())
}

func (d Asset) BusName() string {
	return "MockBus"
}

func (d Asset) UpdateStatus() {
	d.publisher.Publish(msg.Status, AssertedStatus())
}

func (d Asset) UpdateConfig() {
	d.publisher.Publish(msg.Config, AssertedConfig())
}

type Control struct {
	kW   float64
	kVAR float64
}

func (c Control) KW() float64 {
	return c.kW
}

func (c Control) KVAR() float64 {
	return c.kVAR
}

type Status struct {
	machine MachineStatus
}

type Config struct {
	Name string
	Bus  string
}

type MachineStatus struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volt                 float64 `json:"Volt"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	Gridforming          bool    `json:"Gridforming"`
}

func New() Asset {
	pid, _ := uuid.NewUUID()
	publisher := msg.NewPublisher(pid)
	return Asset{pid, publisher, uuid.UUID{}, Control{}}
}

func (s Status) KW() float64 {
	return s.machine.KW
}

func (s Status) KVAR() float64 {
	return s.machine.KVAR
}

func (s Status) RealPositiveCapacity() float64 {
	return s.machine.RealPositiveCapacity
}

func (s Status) RealNegativeCapacity() float64 {
	return s.machine.RealNegativeCapacity
}
