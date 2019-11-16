package asset

import (
	"github.com/google/uuid"
)

// VirtualAsset defines the interface to the virtual assets.
type VirtualAsset interface {
	PID() uuid.UUID
	LinkToBus(<-chan VirtualStatus) <-chan VirtualStatus
}

type VirtualStatus interface {
	KW() float64
	KVAR() float64
	Hz() float64
	Volt() float64
	Gridforming() bool
}
