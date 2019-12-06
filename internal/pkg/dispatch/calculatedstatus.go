package dispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type CalculatedStatus struct{}

type Status struct {
	capacity   Capacity
	renewables Renewables
}

type Capacity struct {
	RealPositiveCapacity float64
	RealNegativeCapacity float64
}

type Renewables struct {
	RE_KW float64
}

func (b CalculatedStatus) updateMemberStatus(msg asset.Msg, memberStatus map[uuid.UUID]Status) map[uuid.UUID]Status {
	status := memberStatus[msg.PID()]
	switch p := msg.Payload().(type) {
	case asset.Capacity:
		status.capacity = Capacity{
			RealPositiveCapacity: p.RealNegativeCapacity(),
			RealNegativeCapacity: p.RealNegativeCapacity(),
		}
	case asset.Renewable:
		status.renewables = Renewables{
			RE_KW: p.RE_KW(),
		}
	}
	memberStatus[msg.PID()] = status

	return memberStatus
}

func (b CalculatedStatus) updateBusStatus(memberStatus map[uuid.UUID]Status) Status {
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
