package msg

import (
	"sync"

	"github.com/google/uuid"
)

type Publisher interface {
	Subscribe(pid uuid.UUID, topic Topic) <-chan Msg
	Unsubscribe(pid uuid.UUID)
}

type PubSub struct {
	mux   sync.RWMutex
	owner uuid.UUID
	subs  map[Topic]map[uuid.UUID]chan<- Msg
}

func NewPublisher(pid uuid.UUID) *PubSub {
	subs := map[Topic]map[uuid.UUID]chan<- Msg{}
	subs[Status] = make(map[uuid.UUID]chan<- Msg)
	subs[Config] = make(map[uuid.UUID]chan<- Msg)
	subs[Control] = make(map[uuid.UUID]chan<- Msg)
	p := &PubSub{sync.RWMutex{}, pid, subs}
	return p
}

// Subscribe returns a channel on which the pubisher writes the specified topic
func (p *PubSub) Subscribe(pid uuid.UUID, topic Topic) <-chan Msg {
	p.mux.Lock()
	defer p.mux.Unlock()

	// Subscribers given a buffer of 1 msg
	ch := make(chan Msg, 1)
	p.subs[topic][pid] = ch
	return ch
}

// Unsubscribe closes all channels associated with a PID
func (p *PubSub) Unsubscribe(pid uuid.UUID) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, topic := range p.subs {
		if _, ok := topic[pid]; ok {
			close(topic[pid])
			delete(topic, pid)
		}
	}
}

func (p *PubSub) Publish(topic Topic, payload interface{}) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	msg := New(p.owner, payload)
	for _, ch := range p.subs[topic] {
		select {
		case ch <- msg:
		default:
		}
	}
}

type Topic int

const (
	Status Topic = iota
	Control
	Config
	JsonStatus
	JsonControl
	JsonConfig
)

// Msg is
type Msg struct {
	sender  uuid.UUID
	payload interface{}
}

// New is the Msg factor function
func New(sender uuid.UUID, payload interface{}) Msg {
	return Msg{sender, payload}
}

// PID returns the sender's PID
func (m Msg) PID() uuid.UUID {
	return m.sender
}

// Payload returns the message data
func (m Msg) Payload() interface{} {
	return m.payload
}
