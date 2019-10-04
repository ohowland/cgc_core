package bus

import (
	"github.com/google/uuid"
)

// Bus is the interface for power system connectivity graph.
type Bus interface {
	Name() string
	PID() uuid.UUID
	Energized() bool
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
