package acbus

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Relayer interface {
	ReadDeviceStatus() (RelayStatus, error)
}

type ACBus struct {
	mux     *sync.Mutex
	name    string
	pid     uuid.UUID
	relay   Relayer
	members map[uuid.UUID]<-chan asset.AssetStatus
	config  Config
	status  Status
}

type Status struct {
	calc  AggregateCapacity
	relay RelayStatus
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
	ratedVolt float64
	ratedHz   float64
}

func (b ACBus) Name() string {
	return b.name
}

func (b ACBus) PID() uuid.UUID {
	return b.pid
}

func (b *ACBus) AddMember(a asset.Asset) {
	b.members[a.PID()] = a.Subscribe(b.pid)
}

func (b *ACBus) RemoveMember(pid uuid.UUID) {
	delete(b.members, pid)
}

// UpdateStatus requests a physical device read, then updates MachineStatus field.
func (b *ACBus) UpdateStatus() {
	var realPositiveCapacity float64
	var realNegativeCapacity float64
	b.mux.Lock()
	defer b.mux.Unlock()
	for pid, member := range b.members {
		assetStatus, ok := <-member
		if !ok {
			b.RemoveMember(pid)
			continue
		}
		realPositiveCapacity += assetStatus.RealPositiveCapacity()
		realNegativeCapacity += assetStatus.RealNegativeCapacity()
	}
	aggregateCapacity := AggregateCapacity{
		RealPositiveCapacity,
		realNegativeCapacity,
	}

	relayStatus, err := b.relay.ReadDeviceStatus()
	if err != nil {
		// comm fail handling path
		return
	}
	b.status = transform(relayStatus)
}

func transform(relayStatus RelayStatus) Status {
	return Status{
		AggregateCapacity,
		relayStatus,
	}
}

func NewBus(relay Relayer) ACBus {
	id, _ := uuid.NewUUID()
	return ACBus{
		name:    "ACBus",
		pid:     id,
		relay:   relay,
		members: make(map[uuid.UUID]<-chan asset.AssetStatus),
		config: Config{
			ratedVolt: 480.0, // Get from config
			ratedHz:   60.0,  // Get from config
		},
	}
}
