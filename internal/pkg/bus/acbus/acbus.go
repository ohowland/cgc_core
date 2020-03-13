/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members.
*/

package acbus

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// ACBus represents a single electrical AC power system bus.
type ACBus struct {
	mux      *sync.Mutex
	pid      uuid.UUID
	relay    Relayer
	inbox    chan msg.Msg
	members  map[uuid.UUID]chan<- msg.Msg
	dispatch dispatch.Dispatcher
	config   Config
	stop     chan bool
}

// Config represents the static properties of an AC Bus
type Config struct {
	Name      string        `json:"Name"`
	RatedVolt float64       `json:"RatedVolt"`
	RatedHz   float64       `json:"RatedHz"`
	Pollrate  time.Duration `json:"Pollrate"`
}

// New configures and returns an ACbus data structure.
func New(jsonConfig []byte, relay Relayer, dispatch dispatch.Dispatcher) (ACBus, error) {

	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return ACBus{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return ACBus{}, err
	}

	inbox := make(chan msg.Msg)
	stop := make(chan bool)
	members := make(map[uuid.UUID]chan<- msg.Msg)

	return ACBus{
		&sync.Mutex{},
		PID,
		relay,
		inbox,
		members,
		dispatch,
		config,
		stop}, nil
}

// Process aggregates information from members, while members exist.
func (b *ACBus) Process() {
	log.Println("ACBus Process: Loop Started")
	defer close(b.inbox)
	poll := time.NewTicker(b.config.Pollrate * time.Millisecond)
	defer poll.Stop()
loop:
	for {
		select {
		case msg, ok := <-b.inbox:
			if !ok {
				b.removeMember(msg.PID())
				b.dispatch.DropAsset(msg.PID())
			}
			if b.hasMember(msg.PID()) {
				b.dispatch.UpdateStatus(msg)
			}
		case <-poll.C:
			assetControls := b.dispatch.GetControl()
			for pid, control := range assetControls {

				select {
				case b.members[pid] <- msg.New(pid, control):
				default:
				}
			}
		case <-b.stop:
			break loop
		}
	}
	log.Println("ACBus Process: Loop Stopped")
}

// stopProcess terminates the bus. This method is used during a controlled shutdown.
func (b *ACBus) stopProcess() {
	b.mux.Lock()
	defer b.mux.Unlock()
	allPIDs := make([]uuid.UUID, len(b.members))

	for pid := range b.members {
		allPIDs = append(allPIDs, pid)
	}

	for _, pid := range allPIDs {
		delete(b.members, pid)
	}

	b.stop <- true
}

// AddMember links an asset to the bus.
// The bus subscribes to member status updates, and requests control of the asset.
func (b *ACBus) AddMember(a interface{}) {
	b.mux.Lock()
	defer b.mux.Unlock()

	//sub := a.Subscribe(b.pid)
	sub := b.subscribe(a.(msg.Subscriber))

	// create a channel for bus to publish to asset control
	pub := make(chan msg.Msg)

	if ok := a.(asset.Controller).RequestControl(b.pid, pub); ok {
		b.members[a.(asset.Identifier).PID()] = pub
	}

	// aggregate messages from assets subscription into the bus inbox
	go func(sub <-chan msg.Msg, inbox chan<- msg.Msg) {
		for msg := range sub {
			inbox <- msg
		}
	}(sub, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

func (b ACBus) subscribe(s msg.Subscriber) <-chan msg.Msg {
	sub := s.Subscribe(b.pid)
	return sub
}

// removeMember revokes membership of an asset to the bus.
func (b *ACBus) removeMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if ch, ok := b.members[pid]; ok {
		close(ch)
	}
	delete(b.members, pid)

	if len(b.members) < 1 {
		b.stopProcess()
	}
}

// hasMember verifies the membership of an asset.
func (b ACBus) hasMember(pid uuid.UUID) bool {
	return b.members[pid] != nil
}

// Energized returns the state of the bus.
func (b ACBus) Energized() bool {
	hzOk := b.Relayer().Hz() > b.config.RatedHz*0.5
	voltOk := b.Relayer().Volt() > b.config.RatedVolt*0.5
	return hzOk && voltOk
}

// Name is an accessor for the ACBus's configured name.
// Use this only when displaying information to customer.
// PID is used internally.
func (b ACBus) Name() string {
	return b.config.Name
}

// PID is an accessor for the ACBus's process id.
func (b ACBus) PID() uuid.UUID {
	return b.pid
}

// Relayer is an accessor for the assigned bus relay.
func (b ACBus) Relayer() Relayer {
	return b.relay
}
