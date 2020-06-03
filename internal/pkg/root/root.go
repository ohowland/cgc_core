package root

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/bus"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// System is the root node of the control system
type System struct {
	publisher msg.Publisher
	busGraph  *bus.BusGraph
	dispatch  dispatch.Dispatcher
}

func NewSystem(g *bus.BusGraph, d dispatch.Dispatcher) (System, error) {
	pid, err := uuid.NewUUID()
	pub := msg.NewPublisher(pid)

	chStatus, err := g.Subscribe(pid, msg.Status)
	go func(ch <-chan msg.Msg) {
		for m := range ch {
			pub.Forward(m)
		}
	}(chStatus)

	chControl, err := pub.Subscribe(pid, msg.Control)
	g.RequestControl(pid, chControl)
	return System{pub, g, d}, err
}

func (s *System) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return s.publisher.Subscribe(pid, topic)
}
