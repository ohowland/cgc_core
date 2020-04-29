package manualdispatch

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// ManualDispatch is the core datastructure
type ManualDispatch struct {
	mux         *sync.Mutex
	calcStatus  *dispatch.CalculatedStatus
	memberState map[uuid.UUID]State
}

type State struct {
	Status  interface{}
	Control interface{}
	Config  interface{}
}

// New returns a configured ManualDispatch struct
func New(configPath string) (ManualDispatch, error) {
	calcStatus, err := dispatch.NewCalculatedStatus()
	memberState := make(map[uuid.UUID]State)
	return ManualDispatch{
			&sync.Mutex{},
			&calcStatus,
			memberState,
		},
		err
}

// UpdateStatus ...
func (c *ManualDispatch) UpdateStatus(m msg.Msg) {
	c.mux.Lock()
	defer c.mux.Unlock()
	switch m.Topic() {
	case msg.STATUS:
		state := c.memberState[m.PID()]
		state.Status = m.Payload()
		c.memberState[m.PID()] = state
		c.calcStatus.AggregateMemberStatus(m)
	case msg.CONFIG:
		state := c.memberState[m.PID()]
		state.Config = m.Payload()
		c.memberState[m.PID()] = state
		c.calcStatus.AggregateMemberConfig(m)
	default:
		log.Printf("manualdispatch.UpdateStatus(): rejected topic: %d", m.Topic())
	}
}

// DropAsset ...
func (c *ManualDispatch) DropAsset(pid uuid.UUID) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.calcStatus.DropAsset(pid)
	delete(c.memberState, pid)
	return nil
}

// GetControl ...
func (c ManualDispatch) GetControl(pid uuid.UUID) (interface{}, bool) {
	state, ok := c.memberState[pid]
	return state.Control, ok
}

// GetStatus ...
func (c ManualDispatch) GetStatus(pid uuid.UUID) (interface{}, bool) {
	state, ok := c.memberState[pid]
	return state.Status, ok
}

// GetCalcStatus ...
func (c ManualDispatch) GetCalcStatus() map[uuid.UUID]dispatch.Status {
	return c.calcStatus.MemberStatus()
}
