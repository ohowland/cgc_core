package manualdispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

type Core struct {
	mux        *sync.Mutex
	calcStatus *dispatch.CalculatedStatus
	assetState map[uuid.UUID]interface{}
}

func (c *Core) UpdateStatus(msg asset.Msg) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.calcStatus.AggregateMemberStatus(msg)
}

func (c *Core) DropAsset(uuid.UUID) {
	c.mux.Lock()
	defer c.mux.Unlock()

}

func (c Core) GetControl() map[uuid.UUID]interface{} {
	return c.assetState
}
