package bus

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/lib/bus/ac/virtualacbus"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"gotest.tools/assert"
)

type NilBus struct {
	pid uuid.UUID
}

func (b *NilBus) AddMember(n Node) {}

// Subscribe placeholder
func (b *NilBus) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	ch := make(chan msg.Msg)
	return ch, nil
}

// Unsubscribe placeholder
func (b *NilBus) Unsubscribe(pid uuid.UUID) {}

// RequestControl placeholder
func (b *NilBus) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error { return nil }

// PID placeholder
func (b NilBus) PID() uuid.UUID {
	return b.pid
}

func (b NilBus) Name() string {
	return "NilBus"
}

// BEGIN --- Graph Tests

func TestNewGraph(t *testing.T) {
	g, err := NewGraph()
	assert.NilError(t, err)
	assert.Assert(t, g.PID() != uuid.UUID{})
}

func TestAddNode(t *testing.T) {
	g, _ := NewGraph()
	bus1 := NilBus{}

	err := g.AddNode(&bus1)
	assert.NilError(t, err)

	_, ok := g.adjacentcyList[&bus1]
	assert.Assert(t, ok, "Node not found in Graph")
}

func TestAddMultipleNodes(t *testing.T) {
	g, _ := NewGraph()
	bus1 := NilBus{}
	bus2 := NilBus{}

	err := g.AddNode(&bus1)
	assert.NilError(t, err)

	err = g.AddNode(&bus2)
	assert.NilError(t, err)

	_, ok := g.adjacentcyList[&bus1]
	assert.Assert(t, ok)

	_, ok = g.adjacentcyList[&bus2]
	assert.Assert(t, ok)

}

func TestRejectDuplicateNode(t *testing.T) {
	g, _ := NewGraph()
	bus1 := &NilBus{}

	err := g.AddNode(bus1)
	assert.NilError(t, err)
	err = g.AddNode(bus1)
	assertError := fmt.Sprintf("node %p already exists in graph.", bus1)
	assert.Error(t, err, assertError)

}

func TestAddDirectedEdge(t *testing.T) {
	g, _ := NewGraph()
	bus1 := NilBus{}
	bus2 := NilBus{}

	g.AddNode(&bus1)
	g.AddNode(&bus2)

	g.AddDirectedEdge(&bus1, &bus2)

	edges1 := g.Edges(&bus1)
	fmt.Printf("bus1 %p -> %p\n", &bus1, edges1[0])

	found := false
	for _, edge := range edges1 {
		if edge == &bus2 {
			found = true
		}
	}
	assert.Assert(t, found, "Directed edge from bus1 to bus2 was not found in bus1's edge list.")

	edges2 := g.Edges(&bus2)
	fmt.Printf("bus2 %p\n", &bus2)
	found = false
	for _, edge := range edges2 {
		if edge == &bus2 {
			found = true
		}
	}
	assert.Assert(t, !found, "(Undirected) Edge fround from bus2 to bus1 found in bus2's edge list.")
}

func TestAddDirectedEdgeMissingStartNode(t *testing.T) {
	g, _ := NewGraph()
	bus1 := NilBus{}
	bus2 := NilBus{}

	// g.AddNode(&bus1) <- Missing Start Node.
	g.AddNode(&bus2)

	err := g.AddDirectedEdge(&bus1, &bus2)
	assertError := fmt.Sprintf("start node %p does not exist in graph.", &bus1)
	assert.Error(t, err, assertError)
}

func TestAddDirectedEdgeMissingEndNode(t *testing.T) {
	g, _ := NewGraph()
	bus1 := NilBus{}
	bus2 := NilBus{}

	g.AddNode(&bus1)
	// g.AddNode(&bus2) <- Missing End Node.

	err := g.AddDirectedEdge(&bus1, &bus2)
	assertError := fmt.Sprintf("end node %p does not exist in graph.", &bus2)
	assert.Error(t, err, assertError)
}

// --- END Graph Tests

// --- BEGIN BusGraph Tests

func TestSetRootBus(t *testing.T) {
	g, _ := NewBusGraph()
	b := NilBus{}

	g.setRootBus(&b)

	assert.Assert(t, g.rootBus.(*NilBus) == &b)
}

func TestAddMember(t *testing.T) {
	g, _ := NewBusGraph()

	bus1, err := virtualacbus.New("../../../config/bus/virtualACBus.json")
	assert.NilError(t, err)

	g.AddMember(&bus1)
}

// --- END BusGraph Tests
