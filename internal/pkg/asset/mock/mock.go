package mock

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
func randDummyStatus() func() DummyStatus {
	return func() DummyStatus {
		rand.Seed(time.Hour.Nanoseconds())
		status := DummyStatus{MachineDummyStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}}
		return status
	}
}

func randDummyControl() func() DummyControl {
	rand.Seed(time.Hour.Nanoseconds())
	control := DummyControl{rand.Float64(), rand.Float64()}
	return func() DummyControl {
		return control
	}
}

func randDummyConfig() func() DummyConfig {
	rand.Seed(time.Hour.Nanoseconds())
	name := fmt.Sprintf("DummyAsset-%d", rand.Int())
	config := DummyConfig{name, "DummyBus"}
	return func() DummyConfig {
		return config
	}
}

var AssertedStatus = randDummyStatus()
var AssertedControl = randDummyControl()
var AssertedConfig = randDummyConfig()

type DummyAsset struct {
	pid          uuid.UUID
	publisher    *msg.PubSub
	controlOwner uuid.UUID
	Control      DummyControl
}

func (d DummyAsset) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	ch, err := d.publisher.Subscribe(pid, topic)
	return ch, err
}

// Unsubscribe pid from all topic broadcasts
func (d DummyAsset) Unsubscribe(pid uuid.UUID) {
	d.publisher.Unsubscribe(pid)
}

func (d *DummyAsset) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	d.controlOwner = pid
	go d.controlHandler(ch)
	return nil
}

// Shutdown asset processes and cleanup resources
func (d *DummyAsset) Shutdown(*sync.WaitGroup) error {
	return nil
}

func (d *DummyAsset) controlHandler(ch <-chan msg.Msg) {
loop:
	for {
		data, ok := <-ch
		if !ok {
			log.Println("DummyAsset controlHandler() stopping")
			break loop
		}
		control, ok := data.Payload().(DummyControl)
		if !ok {
			log.Println("DummyAsset controlHandler() bad type assertion")
			break loop
		}
		d.Control = control
	}
}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d DummyAsset) Name() string {
	return fmt.Sprintf("DummyAsset-%d", rand.Int())
}

func (d DummyAsset) BusName() string {
	return "DummyBus"
}

func (d DummyAsset) UpdateStatus() {
	d.publisher.Publish(msg.Status, AssertedStatus())
}

func (d DummyAsset) UpdateConfig() {
	d.publisher.Publish(msg.Config, AssertedConfig())
}

type DummyControl struct {
	kW   float64
	kVAR float64
}

func (c DummyControl) KW() float64 {
	return c.kW
}

func (c DummyControl) KVAR() float64 {
	return c.kVAR
}

type DummyStatus struct {
	machine MachineDummyStatus
}

type DummyConfig struct {
	Name string
	Bus  string
}

type MachineDummyStatus struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volt                 float64 `json:"Volt"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	Gridforming          bool    `json:"Gridforming"`
}

func NewDummyAsset() DummyAsset {
	pid, _ := uuid.NewUUID()
	publisher := msg.NewPublisher(pid)
	return DummyAsset{pid, publisher, uuid.UUID{}, DummyControl{}}
}

func (s DummyStatus) KW() float64 {
	return s.machine.KW
}

func (s DummyStatus) KVAR() float64 {
	return s.machine.KVAR
}

func (s DummyStatus) RealPositiveCapacity() float64 {
	return s.machine.RealPositiveCapacity
}

func (s DummyStatus) RealNegativeCapacity() float64 {
	return s.machine.RealNegativeCapacity
}
