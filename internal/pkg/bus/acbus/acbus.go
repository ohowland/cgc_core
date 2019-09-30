package acbus

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Relayer interface {
	Volt() float64
	Hz() float64
}

type ACBus struct {
	pid          uuid.UUID
	relay        Relayer
	members      map[uuid.UUID]asset.Asset
	staticConfig StaticConfig
}

type StaticConfig struct {
	ratedVolt float64
	ratedHz   float64
}

func (b ACBus) PID() uuid.UUID {
	return b.pid
}

func (b ACBus) Energized() bool {
	voltThreshold := b.staticConfig.ratedVolt * 0.5
	hzThreshold := b.staticConfig.ratedHz * 0.5
	if b.relay.Hz() > hzThreshold && b.relay.Volt() > voltThreshold {
		return true
	}
	return false
}

func (b *ACBus) AddMember(a asset.Asset) {
	b.members[a.PID()] = a
}

func (b *ACBus) RemoveMember(a asset.Asset) {
	delete(b.members, a.PID())
}

func NewBus(relay Relayer) ACBus {
	id, _ := uuid.NewUUID()
	return ACBus{
		pid:     id,
		relay:   relay,
		members: make(map[uuid.UUID]asset.Asset),
		staticConfig: StaticConfig{
			ratedVolt: 480.0, // Get from config
			ratedHz:   60.0,  // Get from config
		},
	}
}
