package asset

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	Subscribe(uuid.UUID) <-chan msg.Msg
	Unsubscribe(uuid.UUID)
	UpdateStatus()
	RequestControl(uuid.UUID, <-chan msg.Msg) bool
}