package virtualacbus

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"gotest.tools/assert"
)

// randDummyStatus returns a closure for random DummyAsset Status
func randDummyStatus() func() DummyStatus {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	status := DummyStatus{r.Float64(), r.Float64(), r.Float64(), r.Float64(), false}
	return func() DummyStatus {
		return status
	}
}

var assertedStatus = randDummyStatus()

type DummyAsset struct {
	pid     uuid.UUID
	status  DummyStatus
	send    chan<- asset.VirtualStatus
	recieve <-chan asset.VirtualStatus
}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d *DummyAsset) LinkToBus(busIn <-chan asset.VirtualStatus) <-chan asset.VirtualStatus {
	d.recieve = busIn

	busOut := make(chan asset.VirtualStatus)
	d.send = busOut

	return busOut
}

type DummyStatus struct {
	kW          float64
	kVAR        float64
	hz          float64
	volt        float64
	gridforming bool
}

func (v DummyStatus) KW() float64 {
	return v.kW
}
func (v DummyStatus) KVAR() float64 {
	return v.kVAR

}
func (v DummyStatus) Hz() float64 {
	return v.hz
}

func (v DummyStatus) Volt() float64 {
	return v.volt
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

func newVirtualBus() *VirtualACBus {
	configPath := "../acbus_test_config.json"
	bus, err := New(configPath)
	if err != nil {
		panic(err)
	}
	acbus := bus
	vrbus := acbus.Relayer().(*VirtualACBus)
	return vrbus
}

func TestNewVirtualACBus(t *testing.T) {
	configPath := "../acbus_test_config.json"
	bus, err := New(configPath)
	if err != nil {
		t.Fatal(err)
	}
	acbus := bus
	assert.Assert(t, acbus.Name() == "TEST_Virtual Bus")
}

func TestAddMember(t *testing.T) {
	bus := newVirtualBus()

	asset1 := newDummyAsset()
	asset2 := newDummyAsset()

	bus.AddMember(asset1)
	bus.AddMember(asset2)

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID())
	}

	bus.StopProcess()
}

func TestRemoveMember(t *testing.T) {
	bus := newVirtualBus()

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

	bus.RemoveMember(asset2.PID())

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset3.PID())
		assert.Assert(t, pid != asset2.PID())
	}

	bus.StopProcess()
}

func TestProcessOneGridformer(t *testing.T) {
	bus := newVirtualBus()

	asset1 := newDummyAsset()
	asset1.status.gridforming = true

	bus.AddMember(asset1)
	asset1.send <- asset1.status

	time.Sleep(100 * time.Millisecond)
	gridformer := <-asset1.recieve

	assertStatus := assertedStatus()
	assert.Assert(t, gridformer.KW() == 0)
	assert.Assert(t, gridformer.KVAR() == 0)
	assert.Assert(t, gridformer.Hz() == assertStatus.Hz())
	assert.Assert(t, gridformer.Volt() == assertStatus.Volt())

	bus.StopProcess()
	time.Sleep(1000 * time.Millisecond)
}

func TestProcessOneNongridformer(t *testing.T) {
	bus := newVirtualBus()
	asset1 := newDummyAsset()
	asset1.status.gridforming = false
	bus.AddMember(asset1)

	asset1.send <- asset1.status

	time.Sleep(100 * time.Millisecond)
	gridformer := <-asset1.recieve
	assertStatus := assertedStatus()
	assert.Assert(t, gridformer.KW() == -1*assertStatus.KW())
	assert.Assert(t, gridformer.KVAR() == assertStatus.KVAR())
	assert.Assert(t, gridformer.Hz() == 0)
	assert.Assert(t, gridformer.Volt() == 0)

	bus.StopProcess()
}

func TestProcessTwoAssets(t *testing.T) {
	bus := newVirtualBus()
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
	assert.Assert(t, gridformer.KVAR() == assertStatus.KVAR())
	assert.Assert(t, gridformer.Hz() == assertStatus.Hz())
	assert.Assert(t, gridformer.Volt() == assertStatus.Volt())

	bus.StopProcess()
}

func TestReadDeviceStatus(t *testing.T) {
	bus := newVirtualBus()
	asset1 := newDummyAsset()
	asset1.status.gridforming = true
	bus.AddMember(asset1)

	asset1.send <- asset1.status

	time.Sleep(100 * time.Millisecond)
	relayStatus, _ := bus.ReadRelayStatus()

	assertStatus := assertedStatus()
	assert.Assert(t, relayStatus.Hz() == assertStatus.Hz())
	assert.Assert(t, relayStatus.Volt() == assertStatus.Volt())

	bus.StopProcess()
}

/*
func TestBusPowerBalance(t *testing.T) {
	bus := newVirtualBus()

	asset1 := NewDummyAsset()
	asset2 := NewDummyAsset()

	bus.assetObserver <- asset1.asSource()
	bus.assetObserver <- asset2.asSource()
	gridformer := <-bus.busObserver

	assertKwSum := -1 * (asset1.status.kW + asset2.status.kW)
	assertKvarSum := asset1.status.kVAR + asset2.status.kVAR

	assert.Assert(t, gridformer.KW == assertKwSum)
	assert.Assert(t, gridformer.KVAR == assertKvarSum)

	close(bus.assetObserver)
}

func TestBusPowerBalanceChange(t *testing.T) {
	bus := newVirtualBus()

	gridfollowingAsset1 := NewDummyAsset()
	gridfollowingAsset2 := NewDummyAsset()
	gridformingAsset := NewDummyAsset()

	gridformingAsset.status.gridforming = true

	bus.assetObserver <- gridfollowingAsset1.asSource()
	bus.assetObserver <- gridformingAsset.asSource()
	bus.assetObserver <- gridfollowingAsset2.asSource()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 1; i <= 5; i++ {
		gridfollowingAsset1.status.kW = r.Float64() * 1000
		gridfollowingAsset2.status.kW = r.Float64() * 1000

		bus.assetObserver <- gridfollowingAsset2.asSource()
		bus.assetObserver <- gridformingAsset.asSource()
		bus.assetObserver <- gridfollowingAsset1.asSource()
	}

	time.Sleep(time.Duration(200) * time.Millisecond)

	gridformer := <-bus.busObserver
	assertKwSum := -1 * (gridfollowingAsset1.status.kW + gridfollowingAsset2.status.kW)
	assertKvarSum := gridfollowingAsset1.status.kVAR + gridfollowingAsset2.status.kVAR
	assert.Assert(t, gridformer.KW == assertKwSum)
	assert.Assert(t, gridformer.KVAR == assertKvarSum)

	close(bus.assetObserver)
}
*/
