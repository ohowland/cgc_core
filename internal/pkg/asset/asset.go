package asset

import (
	"github.com/google/uuid"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	Name() string
	UpdateStatus()
	WriteControl()
	DispatchControlHandle() MachineControl
	OperatorControlHandle() MachineControl
}

type PowerReader interface {
	KW() float64
	KVAR() float64
}

type CapacityReader interface {
	RealPositive() float64
	RealNegative() float64
	ReactiveSourcing() float64
	ReactiveSinking() float64
}

type StateReader interface {
	Dispatchable() bool // Capacity is offline, but can brought online.
	Operative() bool    // Capacity is available to the system explicitly or implicitly.
}

type MachineControl interface {
	KWCmd(float64)
	KVARCmd(float64)
	RunCmd(bool)
	GridformCmd(bool)
}
