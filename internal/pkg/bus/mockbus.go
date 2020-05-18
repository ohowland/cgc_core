package bus

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type MockBus struct {
	mux            *sync.Mutex
	pid            uuid.UUID
	Members        map[uuid.UUID]Node
	ControlOwner   uuid.UUID
	LastControlMsg msg.Msg
	publisher      *msg.PubSub
}

func NewMockBus() (MockBus, error) {
	pid, _ := uuid.NewUUID()
	m := make(map[uuid.UUID]Node)
	pub := msg.NewPublisher(pid)
	return MockBus{&sync.Mutex{}, pid, m, uuid.UUID{}, msg.Msg{}, pub}, nil
}

// AddMember links the asset parameter to the bus. Asset update status and update
// configuration events will publish to the bus
func (b *MockBus) AddMember(n Node) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Members[n.PID()] = n
	return nil
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
	go b.controlHandler(ch)
	return nil
}

func (b *MockBus) controlHandler(ch <-chan msg.Msg) {
	ctrlMsg, ok := <-ch
	if !ok {
		return
	}
	b.LastControlMsg = ctrlMsg
}

// PID process id
func (b MockBus) PID() uuid.UUID {
	return b.pid
}

func (b MockBus) Name() string {
	return "MockBus"
}
