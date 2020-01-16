package dispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Dispatcher interface {
	UpdateStatus(asset.Msg)
	DropAsset(uuid.UUID)
	GetControl() map[uuid.UUID]asset.Msg
}
