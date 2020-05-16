package dispatch

import (
	"github.com/ohowland/cgc/internal/pkg/bus/ac"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type Dispatcher interface {
	msg.Publisher
}
