package manualdispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

// ManualDispatch is the core datastructure
type ManualDispatch struct {
	mux           *sync.Mutex
	calcStatus    *dispatch.CalculatedStatus
	memberControl map[uuid.UUID]interface{}	
}

// New returns a configured ManualDispatch struct
func New(configPath string) (ManualDispatch, error) {
	calcStatus, err := dispatch.NewCalculatedStatus()
	memberControl := make(map[uuid.UUID]interface{})
	return ManualDispatch{
			&sync.Mutex{},
			&calcStatus,
			memberControl,
		},
		err
}

// UpdateStatus ...
func (c *ManualDispatch) UpdateStatus(msg asset.Msg) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.calcStatus.AggregateMemberStatus(msg)
}

// DropAsset ...
func (c *ManualDispatch) DropAsset(pid uuid.UUID) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.calcStatus.DropAsset(pid)
	delete(c.memberControl, pid)
	return nil
}

// GetControl ...
func (c ManualDispatch) GetControl() map[uuid.UUID]interface{} {
	return c.memberControl
}

// GetMemberStatus
func (c ManualDispatch) MemberStatus() map[uuid.UUID]dispatch.Status {
	return c.calcStatus.MemberStatus()
}

func (c *ManualDispatch) PostStatus()