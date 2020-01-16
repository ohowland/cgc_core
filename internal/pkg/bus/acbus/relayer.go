package acbus

// Relayer is the interface for a bus relayer.
type Relayer interface {
	ReadDeviceStatus() (RelayStatus, error)
}

// RelayStatus is the interface for data structures to report bus status.
type RelayStatus interface {
	Hz() float64
	Volt() float64
}

// EmptyRelayStatus is a data structure for error/uninitialized/unknown status.
type EmptyRelayStatus struct{}

// Hz is an accessor for relayed frequency.
func (s EmptyRelayStatus) Hz() float64 { return 0 }

// Volt is an accessor for relayed voltage.
func (s EmptyRelayStatus) Volt() float64 { return 0 }
