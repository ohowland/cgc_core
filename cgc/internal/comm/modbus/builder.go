package comm

// Builder for ModbusPoller
type Builder struct {
	resource string
	config   PollerConfig
}
