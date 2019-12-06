package acbus

type Relayer interface {
	ReadDeviceStatus() (RelayStatus, error)
}

type RelayStatus interface {
	Hz() float64
	Volt() float64
}

type EmptyRelayStatus struct{}

func (s EmptyRelayStatus) Hz() float64   { return 0 }
func (s EmptyRelayStatus) Volt() float64 { return 0 }
