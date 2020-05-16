package bus

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type Graph struct {
	adjacentcyList map[Bus][]Node
}

type Bus interface {
	Node
	AddMember(Node)
}

type Node interface {
	msg.Publisher
	RequestControl(uuid.UUID, <-chan msg.Msg) error
	PID() uuid.UUID
}

/*
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

type BusGraph interface {
	AssetMembers() map[uuid.UUID]asset.Asset
	BusMembers() map[uuid.UUID]Bus
}
*/
