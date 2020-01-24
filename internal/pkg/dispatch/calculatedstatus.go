package dispatch

import (
	"log"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type CalculatedStatus struct {
	mux          *sync.Mutex
	memberStatus map[uuid.UUID]Status
}

type Status struct {
	power    power
	capacity capacity
}

type capacity struct {
	realPositiveCapacity float64
	realNegativeCapacity float64
}

type power struct {
	kW   float64
	kVAR float64
}

func (s Status) KW() float64 {
	return s.power.kW
}

func (s Status) KVAR() float64 {
	return s.power.kVAR
}

func (s Status) RealPositiveCapacity() float64 {
	return s.capacity.realPositiveCapacity
}

func (s Status) RealNegativeCapacity() float64 {
	return s.capacity.realNegativeCapacity
}

func NewCalculatedStatus() (CalculatedStatus, error) {
	memberStatus := make(map[uuid.UUID]Status)
	return CalculatedStatus{&sync.Mutex{}, memberStatus}, nil
}

func (b *CalculatedStatus) AggregateMemberStatus(msg asset.Msg) {
	status := b.memberStatus[msg.PID()]
	switch p := msg.Payload().(type) {
	case asset.Status:
		status.capacity = capacity{
			realPositiveCapacity: p.RealPositiveCapacity(),
			realNegativeCapacity: p.RealNegativeCapacity(),
		}
		status.power = power{
			kW:   p.KW(),
			kVAR: p.KVAR(),
		}
	default:
		log.Println("Calculated Status Rejected:", reflect.TypeOf(msg.Payload()))
	}
	b.memberStatus[msg.PID()] = status
}

func (b CalculatedStatus) updateBusStatus(msg asset.Msg, memberStatus map[uuid.UUID]Status) Status {
	return Status{
		capacity: aggregateCapacity(memberStatus),
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
		realPositiveCapacity += status.capacity.realPositiveCapacity
		realNegativeCapacity += status.capacity.realNegativeCapacity
	}
	return capacity{
		realPositiveCapacity,
		realNegativeCapacity,
	}
}

func (c CalculatedStatus) MemberStatus() map[uuid.UUID]Status {
	return c.memberStatus
}
