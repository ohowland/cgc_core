package bus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type Bus interface {
	AddMember(Node)
	Node
}

type Node interface {
	msg.Publisher
	RequestControl(uuid.UUID, <-chan msg.Msg) error
	PID() uuid.UUID
}

type Graph struct {
	pid            uuid.UUID
	adjacentcyList map[Node][]Node
}

func NewGraph() (Graph, error) {
	pid, err := uuid.NewUUID()
	if err != nil {
		return Graph{}, err
	}

	al := make(map[Node][]Node)
	return Graph{pid, al}, nil
}

func (g Graph) PID() uuid.UUID {
	return g.pid
}

func (g *Graph) AddNode(n Node) error {
	if _, exists := g.adjacentcyList[n]; exists {
		err := fmt.Sprintf("node %v already exists in graph.", n.PID())
		return errors.New(err)
	}
	edgeList := make([]Node, 0)
	g.adjacentcyList[n] = edgeList
	return nil
}

func (g *Graph) AddDirectedEdge(n1 Node, n2 Node) error {
	edges1, exists := g.adjacentcyList[n1]
	if !exists {
		err := fmt.Sprintf("start node %v does not exist in graph.", n1.PID())
		return errors.New(err)
	}

	if _, exists := g.adjacentcyList[n1]; !exists {
		err := fmt.Sprintf("end node %v does not exist in graph.", n2.PID())
		return errors.New(err)
	}

	g.adjacentcyList[n1] = append(edges1, n2)
	return nil
}

func (g *Graph) Edges(n Node) []Node {
	if edges, exists := g.adjacentcyList[n]; exists {
		return edges
	}
	return make([]Node, 0)
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
