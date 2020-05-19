package root

import (
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

func NewSystem(bg bus.busGraph, d dispatch.Dispatcher) System {
	pub := msg.NewPublisher()
	return System{pub, bg, d}
}
