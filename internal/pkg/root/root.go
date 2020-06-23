package root

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/bus"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// System is the root node of the control system
type System struct {
	pid       uuid.UUID
	publisher msg.Publisher
	busGraph  *bus.BusGraph
	dispatch  dispatch.Dispatcher
}

func NewSystem(g *bus.BusGraph, d dispatch.Dispatcher) (System, error) {
	pid, err := uuid.NewUUID()
	pub := msg.NewPublisher(pid)

	// subscribe System to the busGraph status
	chStatus, err := g.Subscribe(pid, msg.Status)
	if err != nil {
		panic(err)
	}

	// forward all messages recieved from busGraph to subscribers
	go func(ch <-chan msg.Msg) {
		for m := range ch {
			pub.Forward(m)
		}
	}(chStatus)

	chConfig, err := g.Subscribe(pid, msg.Config)
	if err != nil {
		panic(err)
	}

	go func(ch <-chan msg.Msg) {
		for m := range ch {
			pub.Forward(m)
		}
	}(chConfig)

	system := System{pid, pub, g, nil}

	system.setDispatch(d)

	return system, err
}

func (s *System) setDispatch(d dispatch.Dispatcher) error {

	// subscribe to dispatch's control output
	ch, err := d.Subscribe(s.pid, msg.Control)
	if err != nil {
		return err
	}

	// request control of the root bus for dispatch
	err = s.busGraph.RequestControl(s.pid, ch)
	if err != nil {
		return err
	}

	// subscribe Dispatch to the system status
	chControl, err := s.Subscribe(d.PID(), msg.Status)
	if err != nil {
		panic(err)
	}

	// start the dispatch process
	// TODO: this seems too tightly coupled
	err = d.StartProcess(chControl)
	if err != nil {
		panic(err)
	}

	return nil
}

func (s *System) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return s.publisher.Subscribe(pid, topic)
}

func (s *System) Unsubscribe(pid uuid.UUID) {
	s.publisher.Unsubscribe(pid)
}

func (s System) PID() uuid.UUID {
	return s.pid
}
