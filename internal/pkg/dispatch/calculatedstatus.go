package dispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type CalculatedStatus struct {
	memberStatus map[uuid.UUID]Status
}

type Status struct {
	power    Power
	capacity Capacity
}

type Capacity struct {
	RealPositiveCapacity float64
	RealNegativeCapacity float64
}

type Power struct {
	KW   float64
	KVAR float64
}

func (b *CalculatedStatus) AggregateMemberStatus(msg asset.Msg) {
	status := b.memberStatus[msg.PID()]
	switch p := msg.Payload().(type) {
	case asset.Capacity:
		status.capacity = Capacity{
			RealPositiveCapacity: p.RealNegativeCapacity(),
			RealNegativeCapacity: p.RealNegativeCapacity(),
		}
	case asset.Power:
		status.power = Power{
			KW:   p.KW(),
			KVAR: p.KVAR(),
		}
	default:
	}
	b.memberStatus[msg.PID()] = status
}

func (b CalculatedStatus) updateBusStatus(msg asset.Msg, memberStatus map[uuid.UUID]Status) Status {
	return Status{
		capacity: aggregateCapacity(memberStatus),
	}
}

func aggregateCapacity(memberStatus map[uuid.UUID]Status) Capacity {
	var realPositiveCapacity float64
	var realNegativeCapacity float64
	for _, status := range memberStatus {
		realPositiveCapacity += status.capacity.RealPositiveCapacity
		realNegativeCapacity += status.capacity.RealNegativeCapacity
	}
	return Capacity{
		realPositiveCapacity,
		realNegativeCapacity,
	}
}
