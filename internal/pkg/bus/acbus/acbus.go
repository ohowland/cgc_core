/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members. Members report asset.AssetStatus, which is aggregated by the bus.
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

type ACBus struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	relay       Relayer
	inbox       chan asset.Msg
	members     map[uuid.UUID]chan<- interface{}
	dispatch    dispatch.Dispatcher
	config      Config
	stopProcess chan bool
}

type Config struct {
	Name      string        `json:"Name"`
	RatedVolt float64       `json:"RatedVolt"`
	RatedHz   float64       `json:"RatedHz"`
	Pollrate  time.Duration `json:"Pollrate"`
}

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
	members := make(map[uuid.UUID]chan<- interface{})

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

func (b ACBus) Name() string {
	return b.config.Name
}

func (b ACBus) PID() uuid.UUID {
	return b.pid
}

func (b ACBus) Relayer() Relayer {
	return b.relay
}

func (b *ACBus) AddMember(a asset.Asset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender, assetReceiver := a.Subscribe(b.pid)
	b.members[a.PID()] = assetReceiver

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan interface{}, inbox chan<- asset.Msg) {
		for status := range assetSender {
			inbox <- asset.NewMsg(pid, status)
		}
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

func (b *ACBus) removeMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.members, pid)
}

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
				b.dispatch.DropStatus(msg.PID())
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

func (b ACBus) hasMember(pid uuid.UUID) bool {
	return b.members[pid] != nil
}

// UpdateRelayer requests a physical device read, then updates MachineStatus field.
func (b ACBus) UpdateRelayer() (RelayStatus, error) {
	relayStatus, err := b.relay.ReadDeviceStatus()
	if err != nil {
		return EmptyRelayStatus{}, err
	}
	return relayStatus, nil
}
