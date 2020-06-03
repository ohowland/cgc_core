package bus

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"gotest.tools/assert"
)

var timeout time.Duration = 100 * time.Millisecond

func TestNewMockBus(t *testing.T) {
	_, err := NewMockBus()
	assert.NilError(t, err)
}

func TestMockbusAddMember(t *testing.T) {
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()

	bus1.AddMember(&bus2)

	for _, member := range bus1.Members {
		assert.Assert(t, member.PID() == bus2.PID())
	}
}

func TestMockbusSubscribe(t *testing.T) {
	bus1, _ := NewMockBus()

	pid, _ := uuid.NewUUID()
	ch, _ := bus1.Subscribe(pid, msg.Status)

	rand.Seed(time.Now().UnixNano())
	assertInt := rand.Int()
	bus1.publisher.Publish(msg.Status, assertInt)

	incoming := <-ch
	assert.Equal(t, assertInt, incoming.Payload().(int))

}
func TestMockbusUnsubscribe(t *testing.T) {
	bus1, _ := NewMockBus()

	pid, _ := uuid.NewUUID()
	ch, _ := bus1.Subscribe(pid, msg.Status)

	bus1.Unsubscribe(pid)

	_, ok := <-ch
	assert.Assert(t, !ok)
}
func TestMockbusRequestControl(t *testing.T) {
	bus1, _ := NewMockBus()

	pid, _ := uuid.NewUUID()
	ch := make(chan msg.Msg)
	bus1.RequestControl(pid, ch)

	assert.Equal(t, pid, bus1.ControlOwner)

	rand.Seed(time.Now().UnixNano())
	assertMsg := msg.New(pid, msg.Control, rand.Int())
	ch <- assertMsg

	time.Sleep(timeout)
	assert.Equal(t, bus1.LastControlMsg, assertMsg)
}
