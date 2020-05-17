package bus

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

func TestNewGraph(t *testing.T) {
	g, err := NewGraph()
	assert.NilError(t, err)
	assert.Assert(t, g.PID() != uuid.UUID{})
}

func TestAddNode(t *testing.T) {
	g, _ := NewGraph()
	bus1, _ := NewMockBus()

	err := g.AddNode(&bus1)
	assert.NilError(t, err)

	_, ok := g.adjacentcyList[&bus1]
	assert.Assert(t, ok, "Node not found in Graph")
}

func TestAddMultipleNodes(t *testing.T) {
	g, _ := NewGraph()
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()

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
	bus1, _ := NewMockBus()

	err := g.AddNode(&bus1)
	assert.NilError(t, err)
	err = g.AddNode(&bus1)
	assertError := fmt.Sprintf("node %v already exists in graph.", bus1.PID())
	assert.Error(t, err, assertError)

}

func TestAddDirectedEdge(t *testing.T) {
	g, _ := NewGraph()
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()

	g.AddNode(&bus1)
	g.AddNode(&bus2)

	g.AddDirectedEdge(&bus1, &bus2)

	edges1 := g.Edges(&bus1)

	found := false
	for _, edge := range edges1 {
		if edge == &bus2 {
			found = true
		}
	}
	assert.Assert(t, found)

	edges2 := g.Edges(&bus2)

	found = false
	for _, edge := range edges2 {
		if edge == &bus2 {
			found = true
		}
	}
	assert.Assert(t, !found)
}
