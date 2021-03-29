/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members.
*/

package dc

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
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
}

type dynamicConfig struct {
	MemberAssets map[uuid.UUID]member `json:"MemberAssets"`
	MemberBuses  map[uuid.UUID]member `json:"MemberBuses"`
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

	// members are bucketed into buses and assets.
	// this seperation is utilized when a bus recieves
	// a message for an asset that it does not own.
	switch member.node.(type) {
	case bus.Bus:
		b.config.Dynamic.MemberBuses[a.PID()] = member
	case asset.Asset:
		b.config.Dynamic.MemberAssets[a.PID()] = member
	}

	if len(b.config.Dynamic.MemberAssets)+len(b.config.Dynamic.MemberBuses) == 1 { // if this is the first member, start the bus process.
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
		log.Println(m)
		log.Printf("AC Bus %v: recieved message with no target address", b.PID())
		// TODO: Bad Control Message
		return
	}

	if b.hasMember(m.PID()) {
		b.config.Dynamic.MemberAssets[m.PID()].controller <- m
	} else {
		// forward on to member buses
		for pid := range b.config.Dynamic.MemberBuses {
			b.config.Dynamic.MemberBuses[pid].controller <- m
		}
	}
}

// stopProcess terminates the bus. This method is used during a controlled shutdown.
func (b *Bus) stopProcess() {
	b.mux.Lock()
	defer b.mux.Unlock()

	allAssetPIDs := make([]uuid.UUID, len(b.config.Dynamic.MemberAssets))
	for pid := range b.config.Dynamic.MemberAssets {
		allAssetPIDs = append(allAssetPIDs, pid)
	}

	for _, pid := range allAssetPIDs {
		delete(b.config.Dynamic.MemberAssets, pid)
	}

	allBusPIDs := make([]uuid.UUID, len(b.config.Dynamic.MemberBuses))
	for pid := range b.config.Dynamic.MemberBuses {
		allBusPIDs = append(allBusPIDs, pid)
	}

	for _, pid := range allBusPIDs {
		delete(b.config.Dynamic.MemberBuses, pid)
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

	if member, ok := b.config.Dynamic.MemberAssets[pid]; ok {
		member.node.Unsubscribe(b.PID())
		delete(b.config.Dynamic.MemberAssets, pid)
	} else if member, ok := b.config.Dynamic.MemberBuses[pid]; ok {
		member.node.Unsubscribe(b.PID())
		delete(b.config.Dynamic.MemberBuses, pid)
	}

	if len(b.config.Dynamic.MemberAssets)+len(b.config.Dynamic.MemberBuses) < 1 {
		go b.stopProcess()
	}

	// Propigate change in bus dynamic config
	b.UpdateConfig()
}

// hasMember verifies the membership of an asset.
func (b Bus) hasMember(pid uuid.UUID) bool {
	if _, ok := b.config.Dynamic.MemberAssets[pid]; ok {
		return ok
	}

	_, ok := b.config.Dynamic.MemberBuses[pid]
	return ok
}

// Energized returns the state of the bus.
func (b Bus) Energized() bool {
	voltOk := b.Relayer().Volts() > b.config.Static.RatedVolt*0.5
	return voltOk
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
