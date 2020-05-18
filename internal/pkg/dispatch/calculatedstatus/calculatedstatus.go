package dispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

type CalculatedStatus struct {
	mux          *sync.Mutex
	memberStatus map[uuid.UUID]Status
}

type Status struct {
	Power    power
	Capacity capacity
}

func (s Status) KW() float64                   { return s.Power.kW }
func (s Status) KVAR() float64                 { return s.Power.kVAR }
func (s Status) RealPositiveCapacity() float64 { return s.Capacity.realPositiveCapacity }
func (s Status) RealNegativeCapacity() float64 { return s.Capacity.realNegativeCapacity }

type capacity struct {
	realPositiveCapacity float64
	realNegativeCapacity float64
}

type power struct {
	kW   float64
	kVAR float64
}

func NewCalculatedStatus() (CalculatedStatus, error) {
	memberStatus := make(map[uuid.UUID]Status)
	return CalculatedStatus{&sync.Mutex{}, memberStatus}, nil
}

func (b *CalculatedStatus) AggregateMemberStatus(msg msg.Msg) {
	status := b.memberStatus[msg.PID()]
	switch p := msg.Payload().(type) {
	case asset.Status:
		status.capacity = capacity{
			realPositiveCapacity: p.RealPositiveCapacity(),
			realNegativeCapacity: p.RealNegativeCapacity(),
		}
	}

	if assetStatus, ok := m.Payload().(asset.Capacity); ok {
		status.Capacity = capacity{
			realPositiveCapacity: assetStatus.RealPositiveCapacity(),
			realNegativeCapacity: assetStatus.RealNegativeCapacity(),
		}
	}
	b.memberStatus[m.PID()] = status

}

func (b CalculatedStatus) updateBusStatus(msg msg.Msg, memberStatus map[uuid.UUID]Status) Status {
	return Status{
		Capacity: aggregateCapacity(memberStatus),
	}
}

func (c *CalculatedStatus) DropAsset(pid uuid.UUID) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.memberStatus, pid)
}

func aggregateCapacity(memberStatus map[uuid.UUID]Status) capacity {
	var realPositiveCapacity float64
	var realNegativeCapacity float64
	for _, status := range memberStatus {
		realPositiveCapacity += status.Capacity.realPositiveCapacity
		realNegativeCapacity += status.Capacity.realNegativeCapacity
	}
	return capacity{
		realPositiveCapacity,
		realNegativeCapacity,
	}
}

func (c CalculatedStatus) MemberStatus() map[uuid.UUID]Status {
	return c.memberStatus
}
