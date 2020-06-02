package root

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// System is the root node of the control system
type System struct {
	publisher msg.Publisher
	busGraph  bus.Graph
	dispatch  dispatch.Dispatcher
}

func NewSystem(g *bus.busGraph, d dispatch.Dispatcher) (System, error) {
	pid, err := uuid.NewUUID()
	pub := msg.NewPublisher(pid)
	ch := pub.Subscribe(pid, msg.Control)
	g.RequestControl(ch)
	return System{pub, g, d}, err
}

func (s *System) Subscribe(pid uuid.UUID, msg.Topic) <-chan msg.Msg {
	return s.publisher.Subscribe(pid, msg.Topic)
}