package mockbus

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type MockBus struct {
	mux          *sync.Mutex
	pid          uuid.UUID
	Members      map[uuid.UUID]bus.Node
	ControlOwner uuid.UUID
	Control      <-chan msg.Msg
	publisher    *msg.PubSub
}

func NewMockBus() (MockBus, error) {
	pid, _ := uuid.NewUUID()
	m := make(map[uuid.UUID]bus.Node)
	pub := msg.NewPublisher(pid)
	return MockBus{&sync.Mutex{}, pid, m, uuid.UUID{}, nil, pub}, nil
}

// AddMember links the asset parameter to the bus. Asset update status and update
// configuration events will publish to the bus
func (b *MockBus) AddMember(n bus.Node) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Members[n.PID()] = n
}

// Subscribe returns a channel on which the specified topic is broadcast
func (b *MockBus) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	ch, err := b.publisher.Subscribe(pid, topic)
	return ch, err
}

// Unsubscribe pid from all topic broadcasts
func (b *MockBus) Unsubscribe(pid uuid.UUID) {
	b.publisher.Unsubscribe(pid)
}

// RequestControl assigns a channel parameter to the bus control channel
func (b *MockBus) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	// TODO: previous owner needs to stop. how to enforce?
	b.ControlOwner = pid
	b.Control = ch
	return nil
}

// PID process id
func (b MockBus) PID() uuid.UUID {
	return b.pid
}
