package ac

import (
	"io/ioutil"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/mock"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

func newACBus() Bus {
	configPath := "./ac_test_config.json"
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

func TestNewAcBus(t *testing.T) {
	configPath := "./ac_test_config.json"
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	bus, err := New(jsonConfig, DummyRelay{})
	if err != nil {
		t.Fatal(err)
	}
	acbus := bus
	assert.Assert(t, acbus.Name() == "TEST_Virtual Bus")
}

func TestAddMember(t *testing.T) {
	bus := newACBus()

	asset1 := mock.NewDummyAsset()
	asset2 := mock.NewDummyAsset()

	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	assert.Assert(t, len(bus.members) == 2)
	for pid := range bus.members {
		assert.Assert(t, pid == asset1.PID() || pid == asset2.PID())
	}
}

func TestRemoveMember(t *testing.T) {
	bus := newACBus()

	asset1 := mock.NewDummyAsset()
	asset2 := mock.NewDummyAsset()
	asset3 := mock.NewDummyAsset()

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

	bus.removeMember(asset1.PID())
	bus.removeMember(asset3.PID())

	for range bus.members {
		assert.Assert(t, false) // if members is empty this loop will not run.
	}

}

func TestUpdateMemberStatus(t *testing.T) {
	bus := newACBus()

	asset1 := mock.NewDummyAsset()
	asset2 := mock.NewDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	// assets status is pushed to the bus process, which pushes to dispatch
	// asset.UpdateStatus() initiates the cycle.

	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	ch, err := bus.Subscribe(pid, msg.Status)
	assert.NilError(t, err)

	assertStatus := mock.AssertedStatus()

	asset1.UpdateStatus()
	m := <-ch
	assert.Equal(t, m.Payload().(mock.DummyStatus), assertStatus)

	asset2.UpdateStatus()
	m = <-ch
	assert.Equal(t, m.Payload().(mock.DummyStatus), assertStatus)
}

func TestPushControl(t *testing.T) {
	bus := newACBus()

	asset1 := mock.NewDummyAsset()
	asset2 := mock.NewDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	assertControl := mock.AssertedControl()

	pid, _ := uuid.NewUUID()
	ch := make(chan msg.Msg)
	bus.RequestControl(pid, ch)

	ch <- msg.New(pid, msg.New(asset1.PID(), assertControl))

	time.Sleep(100 * time.Millisecond)
	assert.Assert(t, asset1.Control == assertControl, "Failed: %v != %v", asset1.Control, assertControl)

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

func TestStop(t *testing.T) {
	bus := newACBus()

	asset1 := mock.NewDummyAsset()
	asset2 := mock.NewDummyAsset()
	bus.AddMember(&asset1)
	bus.AddMember(&asset2)

	n := 0
	for range bus.members {
		n++
	}
	assert.Equal(t, n, 2)

	bus.stopProcess()

	n = 0
	for range bus.members {
		n++
	}
	assert.Equal(t, n, 0)

}

func TestHasMember(t *testing.T) {
	bus := newACBus()

	asset1 := mock.NewDummyAsset()
	asset2 := mock.NewDummyAsset()
	asset3 := mock.NewDummyAsset()

	bus.AddMember(&asset1)
	bus.AddMember(&asset3)

	assert.Assert(t, bus.hasMember(asset1.PID()))
	assert.Assert(t, !bus.hasMember(asset2.PID()))
	assert.Assert(t, bus.hasMember(asset3.PID()))
}
