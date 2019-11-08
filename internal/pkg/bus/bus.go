package bus

import (
	"github.com/google/uuid"
)

// Bus is the interface for power system connectivity graph.
type Bus interface {
	PID() uuid.UUID
	Name() string
}

type BusGraph struct {
	rootNode       Bus
	adjacentcyList map[Bus][]Bus
}

// NewBusGraph builds a graph datastructure of the buses
func NewBusGraph(buses map[uuid.UUID]Bus) BusGraph {
	var busList []Bus
	for _, bus := range buses {
		busList = append(busList, bus)
	}

	busAdjList := make(map[Bus][]Bus)
	busAdjList[busList[0]] = busList[1:]
	return BusGraph{
		busList[0],
		busAdjList,
	}
}

/*
type BusGroup interface {
	PID() uuid.UUID
	BusGraph() BusGraph
	Energized() bool
}

type BusGraph interface {
	AssetMembers() map[uuid.UUID]asset.Asset
	BusMembers() map[uuid.UUID]Bus
}


TODO:

1. the bus object constructs a bus graph.
2. the virtual system model class should request members from the bus,
poll those members for load information then report the swing load to the
gridformer on the bus.
3.
*/
