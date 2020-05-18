package bus

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
)

// Bus defines interface for a power system bus. Buses are nodes in the graph of
// the power system that can be internal or leaves. As opposed to Assets, which
// are contrainted to leaves only.
type Bus interface {
	AddMember(Node)
	Node
}

// BusGraph is the graph representation of the power system bus.
type BusGraph struct {
	rootBus Bus
	graph   *Graph
}

func NewBusGraph() (BusGraph, error) {
	g, err := NewGraph()
	return BusGraph{&NilBus{}, &g}, err
}

func BuildBusGraph(root Bus, bm map[uuid.UUID]Bus, am map[uuid.UUID]asset.Asset) (BusGraph, error) {
	g, err := NewBusGraph()
	g.setRootBus(root)

	err = g.attachBuses(bm)
	err = g.attachAssets(am)

	return g, err
}

func (bg *BusGraph) setRootBus(b Bus) {
	bg.rootBus = b
}

func (bg *BusGraph) AddMember(n Node) error {
	switch v := n.(type) {
	case Bus:
		if _, ok := bg.rootBus.(*NilBus); ok {
			bg.setRootBus(v)
		} else {
			bg.rootBus.AddMember(v)
		}
	case asset.Asset:

	}

	return nil
}
