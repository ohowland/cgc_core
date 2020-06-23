package bus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// Bus defines interface for a power system bus. Buses are nodes in the graph of
// the power system that can be internal or leaves. As opposed to Assets, which
// are contrainted to leaves only.
type Bus interface {
	AddMember(Node) error
	msg.Publisher
	Controller
	Config
}

type Controller interface {
	UpdateConfig()
	RequestControl(uuid.UUID, <-chan msg.Msg) error
}

type Config interface {
	Name() string
	PID() uuid.UUID
}

// BusGraph is the graph representation of the power system bus.
type BusGraph struct {
	rootBus Bus
	graph   *Graph
}

// NewBusGraph returns and empty BusGraph object
func NewBusGraph() (BusGraph, error) {
	g, err := NewGraph()
	return BusGraph{nil, &g}, err
}

func (bg BusGraph) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return bg.rootBus.Subscribe(pid, topic)
}

func (bg BusGraph) Unsubscribe(pid uuid.UUID) {
	bg.rootBus.Unsubscribe(pid)
}

func (bg BusGraph) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	return bg.rootBus.RequestControl(pid, ch)
}

// BuildBusGraph returns a network graph of buses in assets.
// @param root: the node attached to the dispatch system
func BuildBusGraph(root Bus, buses map[uuid.UUID]Bus, assets map[uuid.UUID]asset.Asset) (BusGraph, error) {
	g, err := NewBusGraph()
	if err != nil {
		return BusGraph{}, err
	}

	g.AddMember(root)

	busesCopy := make(map[uuid.UUID]Bus)
	for k, v := range buses {
		busesCopy[k] = v
	}

	// remove root from bus map
	delete(busesCopy, root.PID())

	for _, bus := range busesCopy {
		err = g.AddMember(bus)
		if err != nil {
			return BusGraph{}, err
		}
	}

	for _, asset := range assets {
		err = g.AddMember(asset)
		if err != nil {
			return BusGraph{}, err
		}
	}

	return g, err
}

func (bg *BusGraph) setRootBus(b Bus) {
	bg.rootBus = b
}

// AddMember inserts a node into the network graph.
func (bg *BusGraph) AddMember(n Node) error {
	switch node := n.(type) {
	case Bus:
		bg.graph.AddNode(node)

		if bg.rootBus == nil {
			bg.setRootBus(node)
		} else {
			bg.graph.AddDirectedEdge(bg.rootBus, node)
			bg.rootBus.AddMember(node) // link bus to bus
		}

	case asset.Asset:
		bus, err := bg.findAssetBus(node)
		if err != nil {
			return err
		}

		err = bg.graph.AddNode(node)
		if err != nil {
			return err
		}

		err = bg.graph.AddDirectedEdge(bus, node)
		if err != nil {
			return err
		}

		err = bus.AddMember(node)
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

// DumpString returns a string representation of the network graph
func (bg BusGraph) DumpString() {
	bg.graph.DumpString()
}
