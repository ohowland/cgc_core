package acbus

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Relayer interface {
	Volt() float64
	Hz() float64
}

type AssetObserver struct {
	members map[uuid.UUID]asset.Asset
	events  chan Status
}

type ACBus struct {
	pid      uuid.UUID
	relay    Relayer
	observer AssetObserver
	status   Status
}

type Status struct {
	Hz        float64
	Volt      float64
	KW        float64
	KVAR      float64
	Energized bool
}

func (b ACBus) PID() uuid.UUID {
	return b.pid
}
func (b ACBus) AssetMembers() map[uuid.UUID]asset.Asset {
	return b.observer.members
}
func (b ACBus) Energized() bool {
	return b.status.Energized
}

func (b *ACBus) AddMember(a asset.Asset) {
	b.observer.members[a.PID()] = a
}

func (b *ACBus) RemoveMember(a asset.Asset) {
	delete(b.observer.members, a.PID())
}

func NewBus(relay Relayer) ACBus {
	id, _ := uuid.NewUUID()
	return ACBus{
		pid:   id,
		relay: relay,
		observer: AssetObserver{
			members: make(map[uuid.UUID]asset.Asset),
			events:  make(chan Status, 1),
		},
		status: Status{
			Hz:        0.0,
			Volt:      0.0,
			KW:        0.0,
			KVAR:      0.0,
			Energized: false,
		},
	}
}
