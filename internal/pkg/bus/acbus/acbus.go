/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members. Members report asset.AssetStatus, which is aggregated by the bus.
*/

package acbus

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus"
)

type ACBus struct {
	mux     *sync.Mutex
	pid     uuid.UUID
	relay   Relayer
	members map[uuid.UUID]<-chan asset.AssetStatus
	config  Config
	status  Status
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
	b.members[a.PID()] = a.Subscribe(b.pid)
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
	for pid := range b.members {
		delete(b.members, pid)
	}
}

// Process aggregates information from members, while members exist.
func (b *ACBus) Process() {
	agg := make(map[uuid.UUID]asset.AssetStatus)
loop:
	for {
		for pid, member := range b.members {
			select {
			case assetStatus, ok := <-member:
				if !ok {
					b.RemoveMember(pid)
					delete(agg, pid)
					if len(b.members) == 0 { // if there are no members, end the bus process.
						break loop
					}
				} else {
					agg[pid] = assetStatus
				}
			}
			b.status.aggregateCapacity = aggregateCapacity(agg)
		}
	}
}

func aggregateCapacity(agg map[uuid.UUID]asset.AssetStatus) AggregateCapacity {
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

func New(jsonConfig []byte, relay Relayer) (bus.Bus, error) {

	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return ACBus{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return ACBus{}, err
	}

	members := make(map[uuid.UUID]<-chan asset.AssetStatus)

	return ACBus{&sync.Mutex{}, PID, relay, members, config, Status{}}, nil
}

type Relayer interface {
	ReadRelayStatus() (RelayStatus, error)
}

type RelayStatus struct {
	hz   float64
	volt float64
}

func (s RelayStatus) Hz() float64 {
	return s.hz
}

func (s RelayStatus) Volt() float64 {
	return s.volt
}

func NewRelayStatus(hz float64, volt float64) RelayStatus {
	return RelayStatus{hz: hz, volt: volt}
}

// UpdateRelayer requests a physical device read, then updates MachineStatus field.
func (b *ACBus) UpdateRelayer() (RelayStatus, error) {
	relayStatus, err := b.relay.ReadRelayStatus()
	if err != nil {
		// comm fail handling path
		return RelayStatus{}, nil
	}
	return relayStatus, nil
}
