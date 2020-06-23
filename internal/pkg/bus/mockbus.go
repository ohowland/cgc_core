package bus

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

type MockBus struct {
	mux            *sync.Mutex
	pid            uuid.UUID
	LastControlMsg msg.Msg
	publisher      *msg.PubSub
	config         MockBusConfig
}

type MockBusConfig struct {
	static  StaticConfig
	dynamic DynamicConfig
}

// Config represents the static properties of an AC Bus
type StaticConfig struct {
	Name      string  `json:"Name"`
	RatedVolt float64 `json:"RatedVolt"`
	RatedHz   float64 `json:"RatedHz"`
}

type DynamicConfig struct {
	Members      map[uuid.UUID]Node `json:"Members"`
	ControlOwner uuid.UUID          `json:"ControlOwner"`
}

func NewMockBus() (MockBus, error) {
	pid, _ := uuid.NewUUID()
	pub := msg.NewPublisher(pid)

	config := MockBusConfig{
		StaticConfig{
			Name:      "Bus",
			RatedVolt: 480,
			RatedHz:   60,
		},
		DynamicConfig{
			Members:      make(map[uuid.UUID]Node),
			ControlOwner: uuid.UUID{},
		},
	}

	return MockBus{&sync.Mutex{}, pid, msg.Msg{}, pub, config}, nil
}

// AddMember links the asset parameter to the bus. Asset update status and update
// configuration events will publish to the bus
func (b *MockBus) AddMember(n Node) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	ch, err := n.Subscribe(b.PID(), msg.Status)
	if err != nil {
		return err
	}
	go func(ch <-chan msg.Msg, pub *msg.PubSub) {
		for msg := range ch {
			pub.Forward(msg)
		}
	}(ch, b.publisher)
	b.config.dynamic.Members[n.PID()] = n

	b.UpdateConfig()
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
	b.config.dynamic.ControlOwner = pid
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

func (b MockBus) UpdateConfig() {
	b.publisher.Publish(msg.Config, b.config)
}
