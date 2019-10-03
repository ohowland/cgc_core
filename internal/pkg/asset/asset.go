package asset

import (
	"github.com/google/uuid"
)

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	PID() uuid.UUID
	UpdateStatus() error
	WriteControl() error
	SetControl(interface{}) error
	Status() interface{}
}

// Device is the interface to read/write a physical component.
// this type of read write will almost certianly have some latency associated with it
// and should not be done as a blocking operation
type Device interface {
	ReadDeviceStatus() (interface{}, error)
	WriteDeviceControl(interface{}) error
}
