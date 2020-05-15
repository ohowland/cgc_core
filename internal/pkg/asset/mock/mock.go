package mock

import (
	"fmt"
	"math/rand"
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
	broadcast    chan<- msg.Msg
	control      <-chan msg.Msg
	controlOwner uuid.UUID
}

func (d *DummyAsset) Subscribe(pid uuid.UUID, topic msg.Topic) <-chan msg.Msg {
	ch := make(chan msg.Msg, 1)
	d.broadcast = ch
	return ch
}

func (d *DummyAsset) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) bool {
	d.control = ch
	d.controlOwner = pid
	return true
}

func (d DummyAsset) Unsubscribe(uuid.UUID) {}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d DummyAsset) Name() string {
	return fmt.Sprintf("DummyAsset-%d", rand.Int())
}

func (d DummyAsset) Bus() string {
	return "DummyBus"
}

func (d DummyAsset) UpdateStatus() {
	status := msg.New(d.pid, AssertedStatus())
	d.broadcast <- status
}

func (d DummyAsset) UpdateConfig() {
	config := msg.New(d.pid, AssertedConfig())
	d.broadcast <- config
}

func (d DummyAsset) RequestContol(uuid.UUID, <-chan msg.Msg) bool {
	return true
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
	ch := make(chan msg.Msg, 1)
	return DummyAsset{pid: uuid.New(), broadcast: ch}
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
