package msg

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"
)

func TestSubscribe(t *testing.T) {
	pidPub, err := uuid.NewUUID()
	assert.NilError(t, err)

	pidSub1, err := uuid.NewUUID()
	assert.NilError(t, err)

	pidSub2, err := uuid.NewUUID()
	assert.NilError(t, err)

	pubsub := NewPublisher(pidPub)
	ch1 := pubsub.Subscribe(pidSub1, Status)
	ch2 := pubsub.Subscribe(pidSub2, Status)

	rand.Seed(time.Now().UnixNano())
	randValue := rand.Float64()

	go func(ch <-chan Msg) {
		t.Log("#1 WAITING")
		incoming := <-ch
		assert.Equal(t, incoming.Payload(), randValue, "First subscriber did not recieve the correct published value")
		t.Log("#1 FINISH")
	}(ch1)

	go func(ch <-chan Msg) {
		t.Log("#2 WAITING")
		incoming := <-ch
		assert.Equal(t, incoming.Payload(), randValue, "Second subscriber did not recieve the correct published value")
		t.Log("#2 FINISH")
	}(ch2)

	time.Sleep(1 * time.Second)
	pubsub.Publish(Status, randValue)
	time.Sleep(1 * time.Second)
}

func TestUnsubscribe(t *testing.T) {}

func TestPublish(t *testing.T) {}
