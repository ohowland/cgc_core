package asset

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// randDummyStatus returns a closure for random DummyAsset Status
func randDummyStatus() func() DummyStatus {
	status := DummyStatus{rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), rand.Float64(), false}
	return func() DummyStatus {
		return status
	}
}

func randDummyControl() func() DummyControl {
	control := DummyControl{rand.Float64(), rand.Float64()}
	return func() DummyControl {
		return control
	}
}

var AssertedStatus = randDummyStatus()
var AssertedControl = randDummyControl()

type DummyAsset struct {
	pid          uuid.UUID
	broadcast    chan<- msg.Msg
	control      <-chan msg.Msg
	controlOwner uuid.UUID
}

func (d *DummyAsset) Subscribe(uuid.UUID) <-chan msg.Msg {
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

func (d DummyAsset) UpdateStatus() {
	status := msg.New(d.pid, AssertedStatus())
	d.broadcast <- status
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
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volt                 float64
	RealPositiveCapacity float64
	RealNegativeCapacity float64
	Gridforming          bool
}

func NewDummyAsset() DummyAsset {
	ch := make(chan msg.Msg, 1)
	return DummyAsset{pid: uuid.New(), broadcast: ch}
}
