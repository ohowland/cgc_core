package dc

import (
	"io/ioutil"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset/mockasset"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

func newDCBus() Bus {
	configPath := "./dc_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	bus, err := New(jsonConfig, NewDummyRelay())
	if err != nil {
		panic(err)
	}
	return bus
}

func TestNewDCBus(t *testing.T) {
	configPath := "./dc_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	bus, err := New(jsonConfig, DummyRelay{})
	if err != nil {
		t.Fatal(err)
	}
	dcbus := bus
	assert.Assert(t, dcbus.Name() == "TEST_Virtual Bus")
}

func TestAddMember(t *testing.T) {
	bus := newDCBus()

	asset1 := mockasset.New()
	asset2 := mockasset.New()

	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	assert.Assert(t, len(bus.config.Dynamic.MemberAssets) == 2)
	for pid := range bus.config.Dynamic.MemberAssets {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID())
	}
}

func TestRemoveMember(t *testing.T) {
	bus := newDCBus()

	asset1 := mockasset.New()
	asset2 := mockasset.New()
	asset3 := mockasset.New()

	bus.AddMember(&asset1)
	bus.AddMember(&asset2)
	bus.AddMember(&asset3)

	assert.Assert(t, len(bus.config.Dynamic.MemberAssets) == 3)
	for pid := range bus.config.Dynamic.MemberAssets {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID() || pid == asset3.PID())
	}

	bus.removeMember(asset2.PID())

	assert.Assert(t, len(bus.config.Dynamic.MemberAssets) == 2)
	for pid := range bus.config.Dynamic.MemberAssets {
		assert.Assert(t, pid == asset1.PID() || pid == asset3.PID())
		assert.Assert(t, pid != asset2.PID())
	}

	bus.removeMember(asset1.PID())
	bus.removeMember(asset3.PID())

	for range bus.config.Dynamic.MemberAssets {
		assert.Assert(t, false) // if members is empty this loop will not run.
	}

}

func TestUpdateMemberStatus(t *testing.T) {
	bus := newDCBus()

	asset1 := mockasset.New()
	asset2 := mockasset.New()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	// assets status is pushed to the bus process, which pushes to dispatch
	// asset.UpdateStatus() initiates the cycle.

	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	ch, err := bus.Subscribe(pid, msg.Status)
	assert.NilError(t, err)

	assertStatus := mockasset.AssertedStatus()

	asset1.UpdateStatus()
	m := <-ch
	assert.Equal(t, m.Payload().(mockasset.Status), assertStatus)

	asset2.UpdateStatus()
	m = <-ch
	assert.Equal(t, m.Payload().(mockasset.Status), assertStatus)
}

func TestPushControl(t *testing.T) {
	bus := newDCBus()

	pid, _ := uuid.NewUUID()
	ch := make(chan msg.Msg)
	bus.RequestControl(pid, ch)

	asset1 := mockasset.New()
	asset2 := mockasset.New()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	assertControl := mockasset.AssertedControl()

	ch <- msg.New(pid, msg.Control, msg.New(asset1.PID(), msg.Control, assertControl))

	time.Sleep(100 * time.Millisecond)
	assert.Assert(t, asset1.Control == assertControl, "Failed: %v != %v", asset1.Control, assertControl)

}

func TestGetRelay(t *testing.T) {
	bus := newDCBus()

	relay := bus.Relayer()

	assertStatus := assertedDummyRelay()

	assert.Assert(t, relay.Volts() == assertStatus.Volts())
}

func TestEnergized(t *testing.T) {
	bus := newDCBus()
	assertStatus := assertedDummyRelay()

	voltOk := assertStatus.Volts() > bus.config.Static.RatedVolt*0.5

	if voltOk {
		assert.Assert(t, bus.Energized() == true)
	} else {
		assert.Assert(t, bus.Energized() == false)
	}
}

func TestStop(t *testing.T) {
	bus := newDCBus()

	asset1 := mockasset.New()
	asset2 := mockasset.New()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	n := 0
	for range bus.config.Dynamic.MemberAssets {
		n++
	}
	assert.Equal(t, n, 2)

	bus.stopProcess()

	n = 0
	for range bus.config.Dynamic.MemberAssets {
		n++
	}
	assert.Equal(t, n, 0)

}

func TestHasMember(t *testing.T) {
	bus := newDCBus()

	asset1 := mockasset.New()
	asset2 := mockasset.New()
	asset3 := mockasset.New()

	bus.AddMember(&asset1)
	bus.AddMember(&asset3)

	assert.Assert(t, bus.hasMember(asset1.PID()))
	assert.Assert(t, !bus.hasMember(asset2.PID()))
	assert.Assert(t, bus.hasMember(asset3.PID()))
}

// TODO: Add tests for member buses as opposed ot member assets
