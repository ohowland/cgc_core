package asset

import (
	"github.com/google/uuid"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	UpdateStatus()
	WriteControl(interface{})
	Subscribe(uuid.UUID) <-chan AssetStatus
	Unsubscribe(uuid.UUID)
}

type AssetStatus interface {
	KW() float64
	KVAR() float64
	RealPositiveCapacity() float64
	RealNegativeCapacity() float64
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

// StateReader interface for access to boolean asset state
type StateReader interface {
	Dispatchable() bool // Capacity is offline, but can brought online.
	Operative() bool    // Capacity is available to the system explicitly or implicitly.
}
