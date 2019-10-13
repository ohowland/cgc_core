package ess

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

type Composite struct {
	pid      uuid.UUID
	members  map[uuid.UUID]asset.Asset
	bus      bus.Bus
	strategy Strategy
	dispatch dispatch.Dispatcher
}

type Strategy interface{}

func NewComposite() Composite {
	return Composite{}
}

func (c *Composite) AddMember(a asset.Asset) {
	c.members[a.PID()] = a
}

func (c *Composite) RemoveMember(pid uuid.UUID) {
	delete(c.members, pid)
}

// aggregate member status

// dispatch control to members
