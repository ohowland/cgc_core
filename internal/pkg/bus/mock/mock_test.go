package mockbus

import (
	"testing"

	"gotest.tools/assert"
)

func TestNewMockBus(t *testing.T) {
	_, err := NewMockBus()
	assert.NilError(t, err)
}

func TestAddMember(t *testing.T) {
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()

	bus1.AddMember(&bus2)

	for _, member := range bus1.Members {
		assert.Assert(t, member.PID() == bus2.PID())
	}
}

func TestSubscribe(t *testing.T)      {}
func TestUnsubscribe(t *testing.T)    {}
func TestRequestControl(t *testing.T) {}
