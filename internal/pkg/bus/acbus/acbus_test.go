package acbus

import (
	"io/ioutil"
	"math/rand"
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

var assertedStatus = randDummyStatus()

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

type DummyDispatch struct{}

func (d DummyDispatch) UpdateStatus(asset.Msg) {}
func (d DummyDispatch) DropAsset(uuid.UUID)    {}
func (d DummyDispatch) GetControl() map[uuid.UUID]asset.Msg {
	return make(map[uuid.UUID]asset.Msg)
}

func newDummyDispatch() dispatch.Dispatcher {
	return DummyDispatch{}
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

func TestProcess(t *testing.T) {
	bus := newACBus()

	asset1 := newDummyAsset()
	bus.AddMember(&asset1)

	asset1.UpdateStatus()

	//assertStatus := assertedStatus()
	time.Sleep(100 * time.Millisecond)
	//assert.Assert(t, bus.status.aggregateCapacity.RealPositiveCapacity == assertStatus.realPositiveCapacity)
	//assert.Assert(t, bus.status.aggregateCapacity.RealNegativeCapacity == assertStatus.realNegativeCapacity)

	bus.StopProcess()
}

/*
func TestAggregateCapacity(t *testing.T) {
	bus := newACBus()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()
	bus.AddMember(asset1)
	bus.AddMember(asset2)

	asset1.UpdateStatus()
	asset2.UpdateStatus()
	assertStatus := assertedStatus()

	time.Sleep(100 * time.Millisecond)
	assert.Assert(t, bus.status.aggregateCapacity.RealPositiveCapacity == 2*assertStatus.realPositiveCapacity)
	assert.Assert(t, bus.status.aggregateCapacity.RealNegativeCapacity == 2*assertStatus.realNegativeCapacity)

	bus.StopProcess()
}
*/

/*
func TestUpdateRelayer(t *testing.T) {
	bus := newACBus()

	err := bus.UpdateRelayer()
	if err != nil {
		t.Fatal(err)
	}

	assertStatus := assertedDummyRelay()

	assert.Assert(t, bus.status.relay.Hz() == assertStatus.Hz())
	assert.Assert(t, bus.status.relay.Volt() == assertStatus.Volt())
}
*/
