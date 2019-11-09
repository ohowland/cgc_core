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
	"github.com/ohowland/cgc/internal/pkg/bus"
)

type Relayer interface {
	ReadDeviceStatus() (RelayStatus, error)
}

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

type RelayStatus struct {
	Hz   float64
	Volt float64
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

// Process aggregates information from members, while members exist.
func (b *ACBus) Process() {
	var agg map[uuid.UUID]asset.AssetStatus
loop:
	for {
		for pid, member := range b.members {
			select {
			case assetStatus, ok := <-member:
				if !ok {
					b.RemoveMember(pid)
					delete(agg, pid)
					continue
				}
				agg[pid] = assetStatus
			default:
				if len(b.members) == 0 { // if there are no members, end the bus process.
					break loop
				}
			}
			b.status.aggregateCapacity = aggregateCapacity(agg)
			time.Sleep(1000 * time.Millisecond)
			log.Printf("Bus %v, Capacity: %v\n", b.config.Name, b.status.aggregateCapacity)
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

// UpdateRelayer requests a physical device read, then updates MachineStatus field.
func (b *ACBus) UpdateRelayer() {
	relayStatus, err := b.relay.ReadDeviceStatus()
	if err != nil {
		// comm fail handling path
		return
	}
	b.status.relay = relayStatus
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
