package virtualacbus

import (
	"log"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

const (
	queueSize = 50
)

type VirtualACBus struct {
	pid              uuid.UUID
	members          map[uuid.UUID]asset.Asset
	observer         chan Source
	connectedSources map[uuid.UUID]Source
	gridformer       Source
	staticConfig     StaticConfig
}

type StaticConfig struct {
	ratedVolt float64
	ratedHz   float64
}

type Source struct {
	PID         uuid.UUID
	Hz          float64
	Volt        float64
	KW          float64
	KVAR        float64
	Gridforming bool
}

func (b VirtualACBus) PID() uuid.UUID {
	return b.pid
}

func (b VirtualACBus) Energized() bool {
	voltThreshold := b.staticConfig.ratedVolt * 0.5
	hzThreshold := b.staticConfig.ratedHz * 0.5
	if b.gridformer.Hz > hzThreshold && b.gridformer.Volt > voltThreshold {
		return true
	}
	return false
}

func (b *VirtualACBus) AddMember(a asset.Asset) {
	b.members[a.PID()] = a
}

func (b *VirtualACBus) RemoveMember(a asset.Asset) {
	delete(b.members, a.PID())
}

func (b VirtualACBus) Observer() chan<- Source {
	return b.observer
}

func (b VirtualACBus) Gridformer() Source {
	kwSum := 0.0
	kvarSum := 0.0
	var gridformer Source
	for _, s := range b.connectedSources {
		if s.Gridforming != true {
			kwSum += s.KW
			kvarSum += s.KVAR
		} else {
			gridformer = s
		}
	}
	gridformer.KW = kwSum
	gridformer.KVAR = kvarSum
	return gridformer
}

func NewVirtalACBus() VirtualACBus {
	id, _ := uuid.NewUUID()
	bus := VirtualACBus{
		pid:      id,
		members:  make(map[uuid.UUID]asset.Asset),
		observer: make(chan Source, queueSize),
		gridformer: Source{
			PID:         id,
			Hz:          0.0,
			Volt:        0.0,
			KW:          0.0,
			KVAR:        0.0,
			Gridforming: true,
		},
		staticConfig: StaticConfig{
			ratedVolt: 480.0, // Get from config
			ratedHz:   60.0,  // Get from config
		},
	}
	go runVirtualSystem(&bus)
	return bus
}

func runVirtualSystem(bus *VirtualACBus) {
	log.Println("[VirtualACBus: Running]")
	for {
		source, ok := <-bus.observer

		if !ok {
			break
		}

		bus.connectedSources[source.PID] = source
		log.Printf("[VirtualACBus: Reported Load %source]\n", source)
	}
}
