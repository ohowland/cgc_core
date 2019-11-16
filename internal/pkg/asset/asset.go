package asset

import (
	"github.com/google/uuid"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	UpdateStatus()
	Subscribe(uuid.UUID) <-chan Status
	Unsubscribe(uuid.UUID)
}

type Status interface {
	KW() float64
	KVAR() float64
}

type Capacity interface {
	RealPositiveCapacity() float64
	RealNegativeCapacity() float64
}

type Renewable interface {
	RE_KW() float64
	RE_KVAR() float64
}
