/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members.
*/

package ac

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

// Bus represents a single electrical AC power system bus.
type Bus struct {
	mux   *sync.Mutex
	pid   uuid.UUID
	relay Relayer
	inbox chan msg.Msg
	//publisher *msg.PubSub
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
func New(jsonConfig []byte, relay Relayer, dispatch dispatch.Dispatcher) (Bus, error) {

	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return Bus{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Bus{}, err
	}

	inbox := make(chan msg.Msg)
	stop := make(chan bool)
	members := make(map[uuid.UUID]chan<- msg.Msg)

	return Bus{
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
func (b *Bus) Process() {
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
			for pid, ch := range b.members {
				ctrl, ok := b.dispatch.GetControl(pid)
				if ok {
					select {
					case ch <- msg.New(pid, ctrl):
					default:
					}
				}
			}
		case <-b.stop:
			break loop
		}
	}
	log.Println("ACBus Process: Loop Stopped")
}

// stopProcess terminates the bus. This method is used during a controlled shutdown.
func (b *Bus) stopProcess() {
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
func (b *Bus) AddMember(a asset.Asset) {
	b.mux.Lock()
	defer b.mux.Unlock()

	// subscribe to asset status broadcast
	subStatus := a.Subscribe(b.PID(), msg.Status)

	// aggregate messages from assets subscription into the bus inbox
	go func(subStatus <-chan msg.Msg, inbox chan<- msg.Msg) {
		for msg := range subStatus {
			inbox <- msg
		}
	}(subStatus, b.inbox)

	// subscribe to asset config broadcast
	subConfig := a.Subscribe(b.PID(), msg.Config)

	go func(subConfig <-chan msg.Msg, inbox chan<- msg.Msg) {
		for msg := range subConfig {
			inbox <- msg
		}
	}(subConfig, b.inbox)

	// request configuration broadcast from device
	a.UpdateConfig()

	// create a channel for bus to publish to asset control
	pubControl := make(chan msg.Msg)
	if ok := b.requestControl(a, pubControl); ok {
		b.members[a.PID()] = pubControl
	}

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

func (b Bus) requestControl(i interface{}, pub chan msg.Msg) bool {
	c := i.(asset.Controller)
	ok := c.RequestControl(b.pid, pub)
	return ok
}

/*
func (b Bus) subscribe(i interface{}) <-chan msg.Msg {
	s := i.(msg.Publisher)
	sub := s.Subscribe(b.pid)
	return sub
}
*/

// removeMember revokes membership of an asset to the bus.
func (b *Bus) removeMember(pid uuid.UUID) {
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
func (b Bus) hasMember(pid uuid.UUID) bool {
	return b.members[pid] != nil
}

// Energized returns the state of the bus.
func (b Bus) Energized() bool {
	hzOk := b.Relayer().Hz() > b.config.RatedHz*0.5
	voltOk := b.Relayer().Volt() > b.config.RatedVolt*0.5
	return hzOk && voltOk
}

// Name is an accessor for the ACBus's configured name.
// Use this only when displaying information to customer.
// PID is used internally.
func (b Bus) Name() string {
	return b.config.Name
}

// PID is an accessor for the ACBus's process id.
func (b Bus) PID() uuid.UUID {
	return b.pid
}

// Relayer is an accessor for the assigned bus relay.
func (b Bus) Relayer() Relayer {
	return b.relay
}
