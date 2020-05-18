package bus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

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
