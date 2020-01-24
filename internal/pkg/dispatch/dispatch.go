package dispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Dispatcher interface {
	UpdateStatus(asset.Msg)
	DropAsset(uuid.UUID) error
	GetControl() map[uuid.UUID]interface{}
}
