package dispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type Dispatcher interface {
	UpdateStatus(msg.Msg)
	DropAsset(uuid.UUID) error
	GetControl() map[uuid.UUID]interface{}
}
