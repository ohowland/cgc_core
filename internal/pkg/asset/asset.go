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

type MachineControl interface {
	KWCmd(float64)
	KVARCmd(float64)
	RunCmd(bool)
	GridformCmd(bool)
}
