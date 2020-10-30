package bus

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// Node types can be used as nodes in the Graph struct.
type Node interface {
	msg.Publisher
	RequestControl(uuid.UUID, <-chan msg.Msg) error
	PID() uuid.UUID
	Name() string
}

// Graph is a struct for organizing the graph network in an adjacentcy list.
type Graph struct {
	pid            uuid.UUID
	adjacentcyList map[Node][]Node
}

// NewGraph returns an initialized empty graph.
func NewGraph() (Graph, error) {
	pid, err := uuid.NewUUID()
	if err != nil {
		return Graph{}, err
	}

	al := make(map[Node][]Node)
	return Graph{pid, al}, nil
}

// PID returns the graph's unique identifier
func (g Graph) PID() uuid.UUID {
	return g.pid
}

// AddNode inserts the node into the graph
func (g *Graph) AddNode(n Node) error {
	if _, exists := g.adjacentcyList[n]; exists {
		err := fmt.Sprintf("node %p already exists in graph.", n)
		return errors.New(err)
	}
	edgeList := make([]Node, 0)
	g.adjacentcyList[n] = edgeList
	return nil
}

// AddDirectedEdge inserts an edge from n1 to n2
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

// Edges returns the edge list for node n
func (g *Graph) Edges(n Node) []Node {
	if edges, exists := g.adjacentcyList[n]; exists {
		return edges
	}
	return make([]Node, 0)
}

// asString returns string representation of the graph
func (g Graph) AsString() {
	for node1, edgelist := range g.adjacentcyList {
		line := fmt.Sprintf("%v: [", node1.Name())
		for i, node2 := range edgelist {
			if i == 0 {
				line = strings.Join([]string{line, node2.Name()}, "")
			} else {
				line = strings.Join([]string{line, node2.Name()}, ", ")
			}
		}
		fmt.Printf("%v]\n", line)
	}
}
