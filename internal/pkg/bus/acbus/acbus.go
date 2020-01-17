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
)

// ACBus represents a single electrical AC power system bus.
type ACBus struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	relay       Relayer
	inbox       chan asset.Msg
	members     map[uuid.UUID]chan<- asset.Msg
	dispatch    dispatch.Dispatcher
	config      Config
	stopProcess chan bool
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

	inbox := make(chan asset.Msg)
	stopProcess := make(chan bool)
	members := make(map[uuid.UUID]chan<- asset.Msg)

	return ACBus{
		&sync.Mutex{},
		PID,
		relay,
		inbox,
		members,
		dispatch,
		config,
		stopProcess}, nil
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

// AddMember links an asset to the bus.
// The bus subscribes to member status updates, and requests control of the asset.
func (b *ACBus) AddMember(a asset.Asset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender := a.Subscribe(b.pid)

	assetReceiver := make(chan asset.Msg)
	if ok := a.RequestControl(b.pid, assetReceiver); ok {
		b.members[a.PID()] = assetReceiver
	}

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan asset.Msg, inbox chan<- asset.Msg) {
		for status := range assetSender {
			inbox <- asset.NewMsg(pid, status)
		}
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

// removeMember revokes membership of an asset to the bus.
func (b *ACBus) removeMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	if ch, ok := b.members[pid]; ok {
		close(ch)
	}
	delete(b.members, pid)

}

// StopProcess terminates the bus. This method is used during a controlled shutdown.
func (b *ACBus) StopProcess() {
	b.mux.Lock()
	defer b.mux.Unlock()
	allPIDs := make([]uuid.UUID, len(b.members))

	for pid := range b.members {
		allPIDs = append(allPIDs, pid)
	}

	for _, pid := range allPIDs {
		delete(b.members, pid)
	}

	b.stopProcess <- true
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
				case b.members[pid] <- control:
				default:
				}
			}
		case <-b.stopProcess:
			break loop
		}
	}
	log.Println("ACBus Process: Loop Stopped")
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
