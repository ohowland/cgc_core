package ac

import (
	"io/ioutil"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

func newACBus() ACBus {
	configPath := "./acbus_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	bus, err := New(jsonConfig, NewDummyRelay(), dispatch.NewDummyDispatch())
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
	bus, err := New(jsonConfig, DummyRelay{}, dispatch.NewDummyDispatch())
	if err != nil {
		t.Fatal(err)
	}
	acbus := bus
	assert.Assert(t, acbus.Name() == "TEST_Virtual Bus")
}

func TestAddMember(t *testing.T) {
	bus := newACBus()

	asset1 := asset.NewDummyAsset()
	asset2 := asset.NewDummyAsset()

	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID())
	}
}

func TestRemoveMember(t *testing.T) {
	bus := newACBus()

	asset1 := asset.NewDummyAsset()
	asset2 := asset.NewDummyAsset()
	asset3 := asset.NewDummyAsset()

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

	asset1 := asset.NewDummyAsset()
	asset2 := asset.NewDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	// assets status is pushed to the bus process, which pushes to dispatch
	// asset.UpdateStatus() initiates the cycle.
	asset1.UpdateStatus()
	asset2.UpdateStatus()
	assertStatus := asset.AssertedStatus()

	time.Sleep(100 * time.Millisecond)

	// check the internals of the mock object DummyDispatch.
	// confirm asset status made it to dispatch.
	d := bus.dispatch.(*dispatch.DummyDispatch)
	asset1Msg := d.AssetStatus[asset1.PID()]
	assert.Assert(t, asset1Msg.Payload().(asset.DummyStatus) == assertStatus)
	assert.Assert(t, asset1Msg.PID() == asset1.PID())

	asset2Msg := d.AssetStatus[asset2.PID()]
	assert.Assert(t, asset2Msg.Payload().(asset.DummyStatus) == assertStatus)
	assert.Assert(t, asset2Msg.PID() == asset2.PID())
}

func TestUpdateDispatcherControl(t *testing.T) {
	bus := newACBus()

	asset1 := asset.NewDummyAsset()
	asset2 := asset.NewDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	asset1.UpdateStatus()
	asset2.UpdateStatus()
	assertControl := asset.AssertedControl()

	time.Sleep(100 * time.Millisecond)

	assetControl := bus.dispatch.GetControl()
	assert.Assert(t, assetControl[asset1.PID()].(asset.DummyControl) == assertControl)
	assert.Assert(t, assetControl[asset2.PID()].(asset.DummyControl) == assertControl)
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
