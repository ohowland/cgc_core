package dispatch

import (
	"github.com/ohowland/cgc_core/internal/pkg/bus/ac"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

type Dispatcher interface {
	msg.Publisher
}
