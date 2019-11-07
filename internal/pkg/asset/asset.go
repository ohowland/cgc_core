package asset

import (
	"github.com/google/uuid"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	UpdateStatus()
	WriteControl()
}

// PowerReporter interface for access to asset real and reactive power
type PowerReporter interface {
	KW() float64
	KVAR() float64
}

// MachineController interface for control of machine state
type MachineController interface {
	KW(float64)
	KVAR(float64)
	Run(bool)
}

// Gridformer interface for control of gridforming state
type Gridformer interface {
	Gridform(bool)
}

// CapacityProvider inferface for access to asset real and reactive capacities
type CapacityProvider interface {
	RealPositive() float64
	RealNegative() float64
	ReactiveSourcing() float64
	ReactiveSinking() float64
}

// StateReader interface for access to boolean asset state
type StateReader interface {
	Dispatchable() bool // Capacity is offline, but can brought online.
	Operative() bool    // Capacity is available to the system explicitly or implicitly.
}
