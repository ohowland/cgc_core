package acbus

import (
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
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

var assertedStatus = randDummyStatus()
var assertedControl = randDummyControl()

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

func (d DummyAsset) UpdateStatus() {
	status := msg.New(d.pid, assertedStatus())
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
	kW                   float64
	kVAR                 float64
	hz                   float64
	volt                 float64
	realPositiveCapacity float64
	realNegativeCapacity float64
	gridforming          bool
}

func (s DummyStatus) KW() float64 {
	return s.kW
}

func (s DummyStatus) KVAR() float64 {
	return s.kVAR
}

func (s DummyStatus) RealPositiveCapacity() float64 {
	return s.realPositiveCapacity
}

func (s DummyStatus) RealNegativeCapacity() float64 {
	return s.realNegativeCapacity
}

func newDummyAsset() DummyAsset {
	ch := make(chan msg.Msg, 1)
	return DummyAsset{pid: uuid.New(), broadcast: ch}
}

type DummyRelay struct {
	hz   float64
	volt float64
}

func (d DummyRelay) Hz() float64 {
	return d.hz
}

func (d DummyRelay) Volt() float64 {
	return d.volt
}

func newDummyRelay() Relayer {
	return assertedDummyRelay()
}

func randDummyRelayStatus() func() DummyRelay {
	status := DummyRelay{rand.Float64(), rand.Float64()}
	return func() DummyRelay {
		return status
	}
}

var assertedDummyRelay = randDummyRelayStatus()

type DummyDispatch struct {
	mux          *sync.Mutex
	PID          uuid.UUID
	assetStatus  map[uuid.UUID]msg.Msg
	assetControl map[uuid.UUID]interface{}
}

func (d *DummyDispatch) UpdateStatus(msg msg.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.assetStatus[msg.PID()] = msg
}

func (d *DummyDispatch) DropAsset(uuid.UUID) error {
	return nil
}

func (d *DummyDispatch) GetControl() map[uuid.UUID]interface{} {
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, Msg := range d.assetStatus {
		d.assetControl[Msg.PID()] = assertedControl()
	}
	return d.assetControl
}

func newDummyDispatch() dispatch.Dispatcher {
	status := make(map[uuid.UUID]msg.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	return &DummyDispatch{&sync.Mutex{}, pid, status, control}
}

func newACBus() ACBus {
	configPath := "./acbus_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	bus, err := New(jsonConfig, newDummyRelay(), newDummyDispatch())
	if err != nil {
		panic(err)
	}
	return bus
}

func TestNewAcBus(t *testing.T) {
	configPath := "./acbus_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	bus, err := New(jsonConfig, DummyRelay{}, newDummyDispatch())
	if err != nil {
		t.Fatal(err)
	}
	acbus := bus
	assert.Assert(t, acbus.Name() == "TEST_Virtual Bus")
}

func TestAddMember(t *testing.T) {
	bus := newACBus()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID())
	}
}

func TestRemoveMember(t *testing.T) {
	bus := newACBus()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()
	asset3 := newDummyAsset()

	bus.AddMember(&asset1)
	bus.AddMember(&asset2)
	bus.AddMember(&asset3)

	assert.Assert(t, len(bus.members) == 3)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID() || pid == asset3.PID())
	}

	bus.removeMember(asset2.PID())

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset3.PID())
		assert.Assert(t, pid != asset2.PID())
	}
}

func TestUpdateDispatcherUpdate(t *testing.T) {
	bus := newACBus()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	// assets status is pushed to the bus process, which pushes to dispatch
	// asset.UpdateStatus() initiates the cycle.
	asset1.UpdateStatus()
	asset2.UpdateStatus()
	assertStatus := assertedStatus()

	time.Sleep(100 * time.Millisecond)

	// check the internals of the mock object DummyDispatch.
	// confirm asset status made it to dispatch.
	dispatch := bus.dispatch.(*DummyDispatch)
	asset1Msg := dispatch.assetStatus[asset1.PID()]
	assert.Assert(t, asset1Msg.Payload().(DummyStatus) == assertStatus)
	assert.Assert(t, asset1Msg.PID() == asset1.PID())

	asset2Msg := dispatch.assetStatus[asset2.PID()]
	assert.Assert(t, asset2Msg.Payload().(DummyStatus) == assertStatus)
	assert.Assert(t, asset2Msg.PID() == asset2.PID())
}

func TestUpdateDispatcherControl(t *testing.T) {
	bus := newACBus()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	asset1.UpdateStatus()
	asset2.UpdateStatus()
	assertControl := assertedControl()

	time.Sleep(100 * time.Millisecond)

	assetControl := bus.dispatch.GetControl()
	assert.Assert(t, assetControl[asset1.PID()].(DummyControl) == assertControl)
	assert.Assert(t, assetControl[asset2.PID()].(DummyControl) == assertControl)
}

func TestGetRelay(t *testing.T) {
	bus := newACBus()

	relay := bus.Relayer()

	assertStatus := assertedDummyRelay()

	assert.Assert(t, relay.Hz() == assertStatus.Hz())
	assert.Assert(t, relay.Volt() == assertStatus.Volt())
}

func TestEnergized(t *testing.T) {
	bus := newACBus()
	assertStatus := assertedDummyRelay()

	hzOk := assertStatus.Hz() > bus.config.RatedHz*0.5
	voltOk := assertStatus.Volt() > bus.config.RatedVolt*0.5

	if hzOk && voltOk {
		assert.Assert(t, bus.Energized() == true)
	} else {
		assert.Assert(t, bus.Energized() == false)
	}
}
