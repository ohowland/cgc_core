/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members. Members report asset.AssetStatus, which is aggregated by the bus.
*/

package acbus

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type ACBus struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	relay       Relayer
	inbox       chan Msg
	members     map[uuid.UUID]bool
	config      Config
	status      Status
	stopProcess chan bool
}

type Status struct {
	aggregateCapacity AggregateCapacity
	relay             RelayStatus
}

type AggregateCapacity struct {
	RealPositiveCapacity float64
	RealNegativeCapacity float64
}

type Config struct {
	Name      string  `json:"Name"`
	RatedVolt float64 `json:"RatedVolt"`
	RatedHz   float64 `json:"RatedHz"`
}

func New(jsonConfig []byte, relay Relayer) (ACBus, error) {

	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return ACBus{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return ACBus{}, err
	}

	busReciever := make(chan Msg)
	stopProcess := make(chan bool)
	members := make(map[uuid.UUID]bool)

	return ACBus{
		&sync.Mutex{},
		PID,
		relay,
		busReciever,
		members,
		config,
		Status{},
		stopProcess}, nil
}

func (b ACBus) Name() string {
	return b.config.Name
}

func (b ACBus) PID() uuid.UUID {
	return b.pid
}

func (b ACBus) Energized() bool {
	return b.status.relay.Hz() > 0.1*b.config.RatedHz &&
		b.status.relay.Volt() > 0.1*b.config.RatedVolt
}

func (b ACBus) Relayer() Relayer {
	return b.relay
}

func (b *ACBus) AddMember(a asset.Asset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender := a.Subscribe(b.pid)
	b.members[a.PID()] = true

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan asset.Status, inbox chan<- Msg) {
		for status := range assetSender {
			inbox <- Msg{pid, status}
		}
		inbox <- Msg{pid, EmptyStatus{}} // on close clear contribution with default status.
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

func (b *ACBus) RemoveMember(pid uuid.UUID) {
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
	defer close(b.inbox)
	log.Println("ACBus Process: Loop Started")
	agg := make(map[uuid.UUID]asset.Status)
loop:
	for {
		select {
		case msg, ok := <-b.inbox:
			if !ok {
				b.RemoveMember(msg.PID())
				delete(agg, msg.PID())
			} else {
				agg = b.updateAggregates(msg, agg)
				b.status.aggregateCapacity = aggregateCapacity(agg)
			}
		case <-b.stopProcess:
			break loop
		}
	}
	log.Println("ACBus Process: Loop Stopped")
}

func (b *ACBus) updateAggregates(msg Msg, agg map[uuid.UUID]asset.Status) map[uuid.UUID]asset.Status {
	b.mux.Lock()
	defer b.mux.Unlock()
	agg[msg.PID()] = msg.Status()
	return agg
}

func aggregateCapacity(agg map[uuid.UUID]asset.Status) AggregateCapacity {

	var realPositiveCapacity float64
	var realNegativeCapacity float64
	for _, assetStatus := range agg {
		realPositiveCapacity += assetStatus.RealPositiveCapacity()
		realNegativeCapacity += assetStatus.RealNegativeCapacity()
	}
	return AggregateCapacity{
		realPositiveCapacity,
		realNegativeCapacity,
	}
}

// UpdateRelayer requests a physical device read, then updates MachineStatus field.
func (b *ACBus) UpdateRelayer() error {
	relayStatus, err := b.relay.ReadRelayStatus()
	log.Println("update relay:", relayStatus)
	if err != nil {
		return err
	}
	b.status.relay = relayStatus
	return nil
}

type Msg struct {
	sender uuid.UUID
	status asset.Status
}

func (v Msg) PID() uuid.UUID {
	return v.sender
}

func (v Msg) Status() asset.Status {
	return v.status
}

type Relayer interface {
	ReadRelayStatus() (RelayStatus, error)
}

type RelayStatus struct {
	hz   float64
	volt float64
}

func NewRelayStatus(hz float64, volt float64) RelayStatus {
	return RelayStatus{hz: hz, volt: volt}
}

func (s RelayStatus) Hz() float64 {
	return s.hz
}

func (s RelayStatus) Volt() float64 {
	return s.volt
}

type EmptyStatus struct{}

func (s EmptyStatus) KW() float64                   { return 0 }
func (s EmptyStatus) KVAR() float64                 { return 0 }
func (s EmptyStatus) RealPositiveCapacity() float64 { return 0 }
func (s EmptyStatus) RealNegativeCapacity() float64 { return 0 }
