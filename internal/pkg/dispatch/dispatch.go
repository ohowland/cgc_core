package dispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

type Dispatcher interface {
	PID() uuid.UUID
	msg.Publisher
	StartProcess(<-chan msg.Msg) error
}
