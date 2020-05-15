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
	Power
	Voltage
	Frequency
	Gridforming
}
