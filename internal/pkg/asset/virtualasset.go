package asset

import (
	"github.com/google/uuid"
)

// VirtualAsset defines the interface to the virtual assets.
type VirtualACAsset interface {
	PID() uuid.UUID
	LinkToBus(<-chan VirtualACStatus) <-chan VirtualACStatus
}

type VirtualACStatus interface {
	RealPower
	ReactivePower
	Voltage
	Frequency
	Gridforming
}

type VirtualDCAsset interface {
	PID() uuid.UUID
	LinkToBus(<-chan VirtualDCStatus) <-chan VirtualDCStatus
}

type VirtualDCStatus interface {
	RealPower
	Voltage
	Gridforming
}
