package msg

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"
)

func TestSubscribe(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)
	pubsub := NewPublisher(pid)

	rand.Seed(time.Now().UnixNano())
	nSubs := rand.Intn(10)
	pids := make([]uuid.UUID, 0)
	chs := make([]<-chan Msg, 0)
	for i := 0; i < nSubs; i++ {
		pid, err = uuid.NewUUID()
		assert.NilError(t, err)
		pids = append(pids, pid)
		ch, err := pubsub.Subscribe(pid, Status)
		assert.NilError(t, err)
		chs = append(chs, ch)
	}

	randValue := rand.Float64()
	var wg sync.WaitGroup
	for i, ch := range chs {
		wg.Add(1)
		go func(ch <-chan Msg, i int, wg *sync.WaitGroup) {
			defer wg.Done()
			incoming := <-ch
			assert.Equal(t, incoming.Payload(), randValue, "Subscriber %v did not recieve the correct published value", i)
		}(ch, i, &wg)
	}

	pubsub.Publish(Status, randValue)
	wg.Wait()
}

func TestUnsubscribe(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)
	pubsub := NewPublisher(pid)

	rand.Seed(time.Now().UnixNano())
	nSubs := rand.Intn(9) + 1
	pids := make([]uuid.UUID, 0)
	chs := make([]<-chan Msg, 0)
	for i := 0; i < nSubs; i++ {
		pid, err = uuid.NewUUID()
		assert.NilError(t, err)
		pids = append(pids, pid)
		ch, err := pubsub.Subscribe(pid, Status)
		assert.NilError(t, err)
		chs = append(chs, ch)
	}

	unsub := rand.Intn(nSubs)
	pubsub.Unsubscribe(pids[unsub])

	var wg sync.WaitGroup
	for i, ch := range chs {
		wg.Add(1)
		go func(ch <-chan Msg, i int, wg *sync.WaitGroup) {
			defer wg.Done()
			_, ok := <-ch
			if !ok {
				assert.Equal(t, i, unsub)
				return
			}
			assert.Assert(t, i != unsub)
		}(ch, i, &wg)
	}

	pubsub.Publish(Status, 0)
	wg.Wait()
}

func TestPublish(t *testing.T) {}
