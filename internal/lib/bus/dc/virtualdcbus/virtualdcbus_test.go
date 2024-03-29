package virtualdcbus

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"gotest.tools/assert"
)

// randDummyStatus returns a closure for random DummyAsset Status
func randDummyStatus() func() DummyStatus {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	status := DummyStatus{r.Float64(), r.Float64(), r.Float64(), false}
	return func() DummyStatus {
		return status
	}
}

var assertedStatus = randDummyStatus()

type DummyAsset struct {
	pid     uuid.UUID
	status  DummyStatus
	send    chan<- asset.VirtualDCStatus
	recieve <-chan asset.VirtualDCStatus
}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d *DummyAsset) LinkToBus(busIn <-chan asset.VirtualDCStatus) <-chan asset.VirtualDCStatus {
	d.recieve = busIn
	busOut := make(chan asset.VirtualDCStatus)
	d.send = busOut

	return busOut
}

func (d *DummyAsset) StopProcess() {
	close(d.send)
}

type DummyStatus struct {
	kW          float64
	kVAR        float64
	volts       float64
	gridforming bool
}

func (v DummyStatus) KW() float64 {
	return v.kW
}

func (v DummyStatus) KVAR() float64 {
	return 0
}

func (v DummyStatus) Volts() float64 {
	return v.volts
}

func (v DummyStatus) Gridforming() bool {
	return v.gridforming
}

func newDummyAsset() *DummyAsset {
	pid, _ := uuid.NewUUID()
	return &DummyAsset{
		pid:    pid,
		status: assertedStatus(),
	}
}

func newVirtualBus() *VirtualDCBus {
	configPath := "../../../../pkg/bus/ac/ac_test_config.json"
	bus, err := New(configPath)
	if err != nil {
		panic(err)
	}
	acbus := bus
	vrbus := acbus.Relayer().(*VirtualDCBus)
	return vrbus
}

func TestNewVirtualDCBus(t *testing.T) {
	configPath := "../../../../pkg/bus/ac/ac_test_config.json"
	bus, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}
	acbus := bus
	assert.Assert(t, acbus.Name() == "TEST_Virtual Bus")
}

func TestAddMember(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	bus.AddMember(asset1)
	bus.AddMember(asset2)

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID())
	}
}

func TestRemoveMember(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()
	asset3 := newDummyAsset()

	bus.AddMember(asset1)
	bus.AddMember(asset2)
	bus.AddMember(asset3)

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

func TestProcessOneGridformer(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset1.status.gridforming = true

	bus.AddMember(asset1)
	asset1.send <- asset1.status

	time.Sleep(100 * time.Millisecond)
	gridformer := <-asset1.recieve

	assertStatus := assertedStatus()
	assert.Assert(t, gridformer.KW() == 0)
	assert.Assert(t, gridformer.Volts() == assertStatus.Volts())
}

func TestProcessOneNongridformer(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset1.status.gridforming = false
	bus.AddMember(asset1)

	asset1.send <- asset1.status

	time.Sleep(100 * time.Millisecond)
	gridformer := <-asset1.recieve
	assertStatus := assertedStatus()
	assert.Assert(t, gridformer.KW() == -1*assertStatus.KW())
	assert.Assert(t, gridformer.Volts() == 0)
}

func TestProcessTwoAssets(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()
	asset1.status.gridforming = false
	asset2.status.gridforming = true
	bus.AddMember(asset1)
	bus.AddMember(asset2)

	asset1.send <- asset1.status
	asset2.send <- asset2.status

	time.Sleep(100 * time.Millisecond)
	gridformer := <-asset2.recieve

	assertStatus := assertedStatus()
	assert.Assert(t, gridformer.KW() == -1*assertStatus.KW())
	assert.Assert(t, gridformer.Volts() == assertStatus.Volts())
}

func TestReadHzVoltStatus(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset1.status.gridforming = true
	bus.AddMember(asset1)

	asset1.send <- asset1.status

	time.Sleep(100 * time.Millisecond)

	assertStatus := assertedStatus()
	assert.Assert(t, bus.Volts() == assertStatus.Volts())
}

func TestProcessMessage(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	bus.AddMember(asset1)
	bus.AddMember(asset2)

	msg1 := msg.New(asset1.PID(), msg.Status, asset1.status)
	msg2 := msg.New(asset2.PID(), msg.Status, asset2.status)

	agg := make(map[uuid.UUID]asset.VirtualDCStatus)
	agg = bus.processMsg(msg1, agg)

	_, ok := agg[asset1.PID()]
	assert.Assert(t, ok)
	assert.Assert(t, agg[asset1.PID()].(DummyStatus) == asset1.status)

	_, ok = agg[asset2.PID()]
	assert.Assert(t, !ok)

	agg = bus.processMsg(msg2, agg)
	assert.Assert(t, agg[asset1.PID()].(DummyStatus) == asset2.status)
	assert.Assert(t, agg[asset2.PID()].(DummyStatus) == asset2.status)

	_, ok = agg[asset1.PID()]
	assert.Assert(t, ok)
	_, ok = agg[asset2.PID()]
	assert.Assert(t, ok)
}

func TestDropAggregateOfRemovedMember(t *testing.T) {
	bus := newVirtualBus()
	defer bus.StopProcess()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	bus.AddMember(asset1)
	bus.AddMember(asset2)

	asset1.send <- asset1.status
	asset2.send <- asset2.status

	time.Sleep(100 * time.Millisecond)
	powerBalance := <-asset1.recieve

	assertKwSum := -1 * (asset1.status.kW + asset2.status.kW)

	assert.Assert(t, powerBalance.KW() == assertKwSum)

	asset1.StopProcess()

	time.Sleep(100 * time.Millisecond)
	powerBalance = <-asset2.recieve

	assertKwSum = -1 * asset2.status.kW
	assert.Assert(t, powerBalance.KW() == assertKwSum)

}

func TestBusPowerBalance(t *testing.T) {
	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	asset1.status.gridforming = false
	asset2.status.gridforming = false

	testAssetMap := make(map[uuid.UUID]asset.VirtualDCStatus)

	testAssetMap[asset1.PID()] = asset1.status
	testAssetMap[asset2.PID()] = asset2.status

	powerBalance := busPowerBalance(testAssetMap)

	assertKwSum := -1 * (asset1.status.kW + asset2.status.kW)

	assert.Assert(t, powerBalance.KW() == assertKwSum)
}

func TestBusPowerBalanceGridformer(t *testing.T) {
	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	asset1.status.gridforming = true
	asset2.status.gridforming = false

	testAssetMap := make(map[uuid.UUID]asset.VirtualDCStatus)

	testAssetMap[asset1.PID()] = asset1.status
	testAssetMap[asset2.PID()] = asset2.status

	gridformer := busPowerBalance(testAssetMap)

	assertKwSum := -1 * (asset2.status.kW)

	assert.Assert(t, gridformer.KW() == assertKwSum)
}
