package acbus

import (
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
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
	broadcast    chan<- asset.Msg
	control      <-chan asset.Msg
	controlOwner uuid.UUID
}

func (d *DummyAsset) Subscribe(uuid.UUID) <-chan asset.Msg {
	ch := make(chan asset.Msg, 1)
	d.broadcast = ch
	return ch
}

func (d *DummyAsset) RequestControl(pid uuid.UUID, ch <-chan asset.Msg) bool {
	d.control = ch
	d.controlOwner = pid
	return true
}

func (d DummyAsset) Unsubscribe(uuid.UUID) {}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d DummyAsset) UpdateStatus() {
	status := asset.NewMsg(d.pid, assertedStatus())
	d.broadcast <- status
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
	ch := make(chan asset.Msg, 1)
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

func (d DummyRelay) ReadDeviceStatus() (RelayStatus, error) {
	return d, nil
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
	assetStatus  map[uuid.UUID]asset.Msg
	assetControl map[uuid.UUID]asset.Msg
}

func (d DummyDispatch) UpdateStatus(msg asset.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.assetStatus[msg.PID()] = msg.Payload().(asset.Msg)
}

func (d DummyDispatch) DropAsset(uuid.UUID) {}

func (d DummyDispatch) GetControl() map[uuid.UUID]asset.Msg {
	resp := asset.NewMsg(d.PID, assertedControl())
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, Msg := range d.assetStatus {
		d.assetControl[Msg.PID()] = resp
	}
	return d.assetControl
}

func newDummyDispatch() dispatch.Dispatcher {
	status := make(map[uuid.UUID]asset.Msg)
	control := make(map[uuid.UUID]asset.Msg)
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

	bus.StopProcess()
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

	bus.StopProcess()
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

	bus.StopProcess()
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
	assert.Assert(t, assetControl[asset1.PID()].Payload().(DummyControl) == assertControl)
	assert.Assert(t, assetControl[asset2.PID()].Payload().(DummyControl) == assertControl)

	bus.StopProcess()
}

func TestUpdateRelayer(t *testing.T) {
	bus := newACBus()

	relay, err := bus.UpdateRelayer()
	if err != nil {
		t.Fatal(err)
	}

	assertStatus := assertedDummyRelay()

	assert.Assert(t, relay.Hz() == assertStatus.Hz())
	assert.Assert(t, relay.Volt() == assertStatus.Volt())
}
