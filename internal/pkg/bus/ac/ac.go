/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members.
*/

package ac

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/bus"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// Bus represents a single electrical AC power system bus.
type Bus struct {
	mux       *sync.Mutex
	pid       uuid.UUID
	relay     Relayer
	publisher *msg.PubSub
	inbox     msgHandler
	config    Config
	stop      chan bool
}

type msgHandler struct {
	status  chan msg.Msg
	control <-chan msg.Msg
}

type member struct {
	node       bus.Node
	controller chan<- msg.Msg
}

// Config contains the static (immutable) and dynamic (mutable) configuration
type Config struct {
	Static  StaticConfig  `json:"Static"`
	Dynamic dynamicConfig `json:"Dynamic"`
}

// StaticConfig represents the static properties of an AC Bus
type StaticConfig struct {
	Name      string  `json:"Name"`
	RatedVolt float64 `json:"RatedVolt"`
	RatedHz   float64 `json:"RatedHz"`
}

type dynamicConfig struct {
	Members      map[uuid.UUID]member `json:"Members"`
	ControlOwner uuid.UUID            `json:"ControlOwner"`
}

// New configures and returns an ACbus data structure.
func New(jsonConfig []byte, relay Relayer) (Bus, error) {

	staticConfig := StaticConfig{}
	err := json.Unmarshal(jsonConfig, &staticConfig)
	if err != nil {
		return Bus{}, err
	}

	dynamicConfig := dynamicConfig{
		make(map[uuid.UUID]member),
		uuid.UUID{},
	}

	pid, err := uuid.NewUUID()
	if err != nil {
		return Bus{}, err
	}

	stop := make(chan bool)
	publisher := msg.NewPublisher(pid)
	MsgHandler := msgHandler{
		make(chan msg.Msg),
		make(<-chan msg.Msg),
	}

	return Bus{
		&sync.Mutex{},
		pid,
		relay,
		publisher,
		MsgHandler,
		Config{
			staticConfig,
			dynamicConfig,
		},
		stop}, nil
}

// Process is the Primary Go Routine for the bus. Process aggregates messages and forwards them to subscribers.
func (b *Bus) Process() {
loop:
	for {
		select {
		case m, ok := <-b.inbox.status:
			if !ok {
				b.removeMember(m.PID())
				continue
			}
			b.publisher.Forward(m)
		case m, ok := <-b.inbox.control:
			if !ok {
				// TODO: Lost Controller
				continue
			}
			b.publishMemberControl(m)
		case <-b.stop:
			break loop
		}
	}
}

// AddMember links the asset parameter to the bus. Asset update status and update
// configuration events will publish to the bus
func (b *Bus) AddMember(a bus.Node) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	member, err := b.newMember(a)
	if err != nil {
		// TODO: Error Handling Path: Failure to add member
		return err
	}
	b.config.Dynamic.Members[a.PID()] = member

	if len(b.config.Dynamic.Members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}

	// Propigate change in bus dynamic config
	b.UpdateConfig()
	return nil
}

// UpdateConfig pushes bus configuration to PubSub network
func (b Bus) UpdateConfig() {
	b.publisher.Publish(msg.Config, b.config)
}

// Subscribe returns a channel on which the specified topic is broadcast
func (b *Bus) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	ch, err := b.publisher.Subscribe(pid, topic)
	return ch, err
}

// Unsubscribe pid from all topic broadcasts
func (b *Bus) Unsubscribe(pid uuid.UUID) {
	b.publisher.Unsubscribe(pid)
}

// RequestControl assigns a channel parameter to the bus control channel
func (b *Bus) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	// TODO: previous owner needs to stop. how to enforce?
	b.config.Dynamic.ControlOwner = pid
	b.inbox.control = ch
	return nil
}

func (b Bus) publishMemberControl(m msg.Msg) {
	// TODO: Control Messages are targeted for an asset.
	m, ok := unwrap(m)
	if !ok {
		log.Printf("AC Bus %v: recieved message with no target address", b.PID())
		// TODO: Bad Control Message
		return
	}

	if b.hasMember(m.PID()) {
		b.config.Dynamic.Members[m.PID()].controller <- m
	} else {
		log.Printf("AC Bus %v: recieved message bound for non-member address %v", b.PID(), m.PID())
	}
}

// stopProcess terminates the bus. This method is used during a controlled shutdown.
func (b *Bus) stopProcess() {
	b.mux.Lock()
	defer b.mux.Unlock()
	allPIDs := make([]uuid.UUID, len(b.config.Dynamic.Members))

	for pid := range b.config.Dynamic.Members {
		allPIDs = append(allPIDs, pid)
	}

	for _, pid := range allPIDs {
		delete(b.config.Dynamic.Members, pid)
	}

	b.stop <- true
}

func (b *Bus) newMember(node bus.Node) (member, error) {
	chIn, err := node.Subscribe(b.PID(), msg.Status)
	if err != nil {
		return member{}, err
	}
	go redirectMsg(chIn, b.inbox.status)

	chIn, err = node.Subscribe(b.PID(), msg.Config)
	if err != nil {
		return member{}, err
	}
	go redirectMsg(chIn, b.inbox.status)

	chOut, err := b.requestControl(node)
	if err != nil {
		return member{}, err
	}

	return member{node, chOut}, nil
}

func (b Bus) requestControl(node bus.Node) (chan<- msg.Msg, error) {
	ch := make(chan msg.Msg)
	err := node.RequestControl(b.pid, ch)
	return ch, err
}

// removeMember revokes membership of an asset to the bus.
func (b *Bus) removeMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if member, ok := b.config.Dynamic.Members[pid]; ok {
		member.node.Unsubscribe(b.PID())
	}
	delete(b.config.Dynamic.Members, pid)

	if len(b.config.Dynamic.Members) < 1 {
		go b.stopProcess()
	}

	// Propigate change in bus dynamic config
	b.UpdateConfig()
}

// hasMember verifies the membership of an asset.
func (b Bus) hasMember(pid uuid.UUID) bool {
	_, ok := b.config.Dynamic.Members[pid]
	return ok
}

// Energized returns the state of the bus.
func (b Bus) Energized() bool {
	hzOk := b.Relayer().Hz() > b.config.Static.RatedHz*0.5
	voltOk := b.Relayer().Volts() > b.config.Static.RatedVolt*0.5
	return hzOk && voltOk
}

// Name is an accessor for the ACBus's configured name.
// Use this only when displaying information to customer.
// PID is used internally.
func (b Bus) Name() string {
	return b.config.Static.Name
}

// PID is an accessor for the ACBus's process id.
func (b Bus) PID() uuid.UUID {
	return b.pid
}

// Relayer is an accessor for the assigned bus relay.
func (b Bus) Relayer() Relayer {
	return b.relay
}

func redirectMsg(in <-chan msg.Msg, out chan<- msg.Msg) {
	for ch := range in {
		out <- ch
	}
}

func unwrap(m msg.Msg) (msg.Msg, bool) {
	unwrapped, ok := m.Payload().(msg.Msg)
	return unwrapped, ok
}
