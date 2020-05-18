package bus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// Bus defines interface for a power system bus. Buses are nodes in the graph of
// the power system that can be internal or leaves. As opposed to Assets, which
// are contrainted to leaves only.
type Bus interface {
	AddMember(Node) error
	Node
}

type Node interface {
	msg.Publisher
	RequestControl(uuid.UUID, <-chan msg.Msg) error
	PID() uuid.UUID
	Name() string
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
		err := fmt.Sprintf("node %p already exists in graph.", n)
		return errors.New(err)
	}
	edgeList := make([]Node, 0)
	g.adjacentcyList[n] = edgeList
	return nil
}

func (g *Graph) AddDirectedEdge(n1 Node, n2 Node) error {
	edges1, exists := g.adjacentcyList[n1]
	if !exists {
		err := fmt.Sprintf("start node %p does not exist in graph.", n1)
		return errors.New(err)
	}

	if _, exists := g.adjacentcyList[n2]; !exists {
		err := fmt.Sprintf("end node %p does not exist in graph.", n2)
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

// BusGraph is the graph representation of the power system bus.
type BusGraph struct {
	rootBus Bus
	graph   *Graph
}

func NewBusGraph() (BusGraph, error) {
	g, err := NewGraph()
	return BusGraph{nil, &g}, err
}

func BuildBusGraph(root Bus, bm map[uuid.UUID]Bus, am map[uuid.UUID]asset.Asset) (BusGraph, error) {
	g, err := NewBusGraph()
	g.setRootBus(root)

	// err = g.attachBuses(bm)
	// err = g.attachAssets(am)

	return g, err
}

func (bg *BusGraph) setRootBus(b Bus) {
	bg.rootBus = b
}

func (bg *BusGraph) AddMember(n Node) error {
	switch v := n.(type) {
	case Bus:
		if bg.rootBus == nil {
			bg.setRootBus(v)
		} else {
			bg.rootBus.AddMember(v)
		}
	case asset.Asset:
		if bg.rootBus == nil {
			return errors.New("root bus undefined: add bus before asset")
		}
		targetBus, err := bg.findAssetBus(v)
		if err != nil {
			return err
		}
		err = targetBus.AddMember(v)
		if err != nil {
			return err
		}
	default:
		return errors.New("node type unsupported by busgraph. interface types bus.Bus or asset.Asset supported")
	}

	return nil
}

func (bg *BusGraph) findAssetBus(a asset.Asset) (Bus, error) {
	for _, node := range bg.nodeList() {
		switch v := node.(type) {
		case Bus:
			if v.Name() == a.BusName() {
				return v, nil
			}
		default:
		}
	}
	err := fmt.Sprintf("graph does not contain target bus %v", a.BusName())
	return nil, errors.New(err)
}

func (bg *BusGraph) nodeList() []Node {
	nodeList := make([]Node, 0)
	for node := range bg.graph.adjacentcyList {
		nodeList = append(nodeList, node)
	}
	return nodeList
}
