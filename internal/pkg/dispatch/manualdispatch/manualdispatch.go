package manualdispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

type Core struct {
	mux              *sync.Mutex
	calculatedStatus dispatch.CalculatedStatus
	assetState       map[uuid.UUID]interface{}
}

func (c Core) UpdateStatus(uuid.UUID, interface{}) {

}

func (c *Core) DropStatus(uuid.UUID) {

}

func (c Core) GetControl() map[uuid.UUID]interface{} {
	return c.assetState
}
