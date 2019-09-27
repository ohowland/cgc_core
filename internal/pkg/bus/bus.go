package asset

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Relayer interface {
	Hz() float64
	Volt() float64
}

type Bus struct {
	id      uuid.UUID
	relay   Relayer
	members map[uuid.UUID]asset.Asset
	status  Status
}

type Status struct {
	Hz        float64 `json:"Hz"`
	Volts     float64 `json:"Volts"`
	Energized bool    `json:"Energized"`
}

/*
TODO:

1. the bus object constructs a bus graph.
2. the virtual system model class should request members from the bus,
poll those members for load information then report the swing load to the
gridformer on the bus.
3.

/*
type Source struct {
	ID          uuid.UUID
	Hz          float64
	Volts       float64
	KW          float64
	KVAR        float64
	Gridforming bool
}

func (a Bus) LoadsChan() chan<- Source {
	return a.comm.sources
}

func (a Bus) GridformChan() <-chan Source {
	return a.comm.gridformer
}

// updateVirtualSystem recieves load information for connected assets,
// and calculates the swing load.
func updateVirtualDevice(dev *Bus, comm Comm) *Bus {
	select {
	case s := <-comm.sources:
		dev.status.ConnectedSources[s.ID] = s
		//log.Printf("[VirtualBus-SystemModel: Reported Load %v]\n", v)
	case comm.gridformer <- dev.gridformingLoad():
		gridformer := dev.gridformingLoad()
		dev.status.Hz = gridformer.Hz
		dev.status.Volts = gridformer.Volts
		log.Printf("[VirtualRelay-Device: Gridformer Load %v]\n", gridformer)
	}
	return dev
}

func (a Bus) gridformingLoad() Source {
	kwSum := 0.0
	kvarSum := 0.0
	var swingMachine Source
	for _, s := range a.status.ConnectedSources {
		if s.Gridforming != true {
			kwSum += s.KW
			kvarSum += s.KVAR
		} else {
			swingMachine = s
		}
	}

	swingMachine.KW = kwSum
	swingMachine.KVAR = kvarSum
	return swingMachine
}
*/
