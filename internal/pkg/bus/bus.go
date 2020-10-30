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

// Subscribe returns a channel on which the root node in the bus graph publishes topics.
func (bg BusGraph) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return bg.rootBus.Subscribe(pid, topic)
}

// Unsubscribe removes the listener's PID from the subscription list.
func (bg BusGraph) Unsubscribe(pid uuid.UUID) {
	bg.rootBus.Unsubscribe(pid)
}

// RequestControl attempts to aquire the control channel for the root node of the
// bus graph.
func (bg BusGraph) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	return bg.rootBus.RequestControl(pid, ch)
}

// BuildBusGraph returns a network graph of buses in assets.
// @param root: the node attached to the dispatch system.
func BuildBusGraph(root Bus, buses map[uuid.UUID]Bus, assets map[uuid.UUID]asset.Asset) (BusGraph, error) {
	g, err := NewBusGraph()
	if err != nil {
		return BusGraph{}, err
	}

	g.AddMember(root)

	buses = dropBus(root, buses)
	for _, bus := range dropBus(root, buses) {
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

// dropBus returns a new map[uuid.UUID]Bus that does not contain the dropped bus
func dropBus(drop Bus, buses map[uuid.UUID]Bus) map[uuid.UUID]Bus {
	busesMutant := make(map[uuid.UUID]Bus)
	for k, v := range buses {
		busesMutant[k] = v
	}

	delete(busesMutant, drop.PID())
	return busesMutant
}

func (bg *BusGraph) setRootBus(b Bus) {
	bg.rootBus = b
}

// AddMember inserts a node into the network graph.
// The first bus added as a member will assume the position of root bus.
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

// AsString prints a string representation of the network graph
func (bg BusGraph) AsString() {
	bg.graph.AsString()
}
