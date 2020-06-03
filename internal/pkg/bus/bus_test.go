package bus

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/asset/mockasset"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"gotest.tools/assert"
)

// BEGIN --- Graph Tests

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
	assertError := fmt.Sprintf("node %p already exists in graph.", &bus1)
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
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()

	// g.AddNode(&bus1) <- Missing Start Node.
	g.AddNode(&bus2)

	err := g.AddDirectedEdge(&bus1, &bus2)
	assertError := fmt.Sprintf("start node %p does not exist in graph.", &bus1)
	assert.Error(t, err, assertError)
}

func TestAddDirectedEdgeMissingEndNode(t *testing.T) {
	g, _ := NewGraph()
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()

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
	bus1, _ := NewMockBus()

	g.setRootBus(&bus1)

	assert.Assert(t, g.rootBus == &bus1)
}

func TestAddBusMember(t *testing.T) {
	g, _ := NewBusGraph()
	bus1, _ := NewMockBus()

	err := g.AddMember(&bus1)
	assert.NilError(t, err)
	assert.Assert(t, g.rootBus.(*MockBus) == &bus1)
}

func TestNodeList(t *testing.T) {
	g, _ := NewBusGraph()
	bus1, _ := NewMockBus()
	bus2, _ := NewMockBus()
	bus3, _ := NewMockBus()

	assertSet := make(map[Node]bool)
	assertSet[&bus1] = true
	assertSet[&bus2] = true
	assertSet[&bus3] = true

	g.AddMember(&bus1)
	g.AddMember(&bus2)
	g.AddMember(&bus3)

	for _, bus := range g.nodeList() {
		_, ok := assertSet[bus]
		assert.Assert(t, ok, "bus added as member to graph not found in nodelist")
	}

}
func TestFindAssetBus(t *testing.T) {
	g, _ := NewBusGraph()
	bus1, _ := NewMockBus()
	asset1 := mockasset.New()

	g.AddMember(&bus1)

	bus, err := g.findAssetBus(&asset1)
	assert.NilError(t, err)
	assert.Assert(t, bus == &bus1)
}

func TestAddAssetMember(t *testing.T) {
	g, _ := NewBusGraph()
	bus1, _ := NewMockBus()
	asset1 := mockasset.New()

	err := g.AddMember(&bus1)
	assert.NilError(t, err)

	err = g.AddMember(&asset1)
	assert.NilError(t, err)

	assertSet := make(map[Node]bool)
	assertSet[&bus1] = true
	assertSet[&asset1] = true

	for _, node := range g.nodeList() {
		_, ok := assertSet[node]
		assert.Assert(t, ok, "bus added as member to graph not found in nodelist")
	}

	_, ok := bus1.Members[asset1.PID()]
	assert.Assert(t, ok, "asset1 is not a member of bus1")
}

func TestBuildBusGraph(t *testing.T) {
	bm := make(map[uuid.UUID]Bus)
	bus1, _ := NewMockBus()
	bm[bus1.PID()] = &bus1

	am := make(map[uuid.UUID]asset.Asset)
	asset1 := mockasset.New()
	am[asset1.PID()] = &asset1

	g, err := BuildBusGraph(&bus1, bm, am)
	assert.NilError(t, err)

	assertSet := make(map[Node]bool)
	assertSet[&bus1] = true
	assertSet[&asset1] = true

	for _, node := range g.nodeList() {
		_, ok := assertSet[node]
		assert.Assert(t, ok, "node added as member to graph but not found in nodelist")
	}

	_, ok := bus1.Members[asset1.PID()]
	assert.Assert(t, ok, "asset1 is not a member of bus1")
}

func TestUpdateStatusBusGraph(t *testing.T) {
	bm := make(map[uuid.UUID]Bus)
	bus1, _ := NewMockBus()
	bm[bus1.PID()] = &bus1

	am := make(map[uuid.UUID]asset.Asset)
	asset1 := mockasset.New()
	am[asset1.PID()] = &asset1

	g, err := BuildBusGraph(&bus1, bm, am)
	assert.NilError(t, err)

	pid, _ := uuid.NewUUID()
	ch, err := g.Subscribe(pid, msg.Status)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ch <-chan msg.Msg, wg *sync.WaitGroup) {
		defer wg.Done()
		in := <-ch
		fmt.Println(in)
	}(ch, &wg)

	g.DumpString()
	asset1.UpdateStatus()
	wg.Wait()
}

// --- END BusGraph Tests
