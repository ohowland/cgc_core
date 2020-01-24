package asset

import (
	"github.com/google/uuid"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	UpdateStatus()
	Subscribe(uuid.UUID) <-chan Msg
	Unsubscribe(uuid.UUID)
	RequestControl(uuid.UUID, <-chan Msg) bool
}

type Status interface {
	KW() float64
	KVAR() float64
	RealPositiveCapacity() float64
	RealNegativeCapacity() float64
}

type Msg struct {
	sender  uuid.UUID
	payload interface{}
}

func NewMsg(sender uuid.UUID, payload interface{}) Msg {
	return Msg{sender, payload}
}

func (v Msg) PID() uuid.UUID {
	return v.sender
}

func (v Msg) Payload() interface{} {
	return v.payload
}
