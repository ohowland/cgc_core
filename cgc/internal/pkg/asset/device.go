package asset

// Device is the interface to read/write a physical component.
// this type of read write will almost certianly have some latency associated with it
// and should not be done as a blocking operation
type Device interface {
	ReadDeviceStatus() (interface{}, error)
	WriteDeviceControl(interface{}) error
}
