package root

import (
	"github.com/ohowland/cgc_core/internal/pkg/bus"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// System is the root node of the control system
type System struct {
	publisher msg.Publisher
	busGraph  bus.Graph
	dispatch  dispatch.Dispatcher
}

func (s *System) SetBusGraph(bus.Graph) {}

func (s *System) SetDispatcher(dispatch.Dispatcher) {}
