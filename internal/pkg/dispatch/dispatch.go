package dispatch

type Dispatcher interface{}

type StateController interface {
	Standby(bool) error
	Gridform(bool) error
}

type PowerController interface {
	KWSetpoint(float64) error
	KVARSetpoint(float64) error
}

type CapacityController interface {
	RealPositiveCapacitySetpoint(float64) error
	RealNegativeCapacitySetpoint(float64) error
	ReactiveSourcingCapacitySetpoint(float64) error
	ReactiveSinkingCapacitySetpoint(float64) error
}

type PowerReporter interface {
	KW() float64
	KVAR() float64
}

type CapacityReporter interface {
	RealPositiveCapacityOperative() float64
	RealNegativeCapacityOperative() float64
	ReactiveSourcingCapacityOperative() float64
	ReactiveSinkingCapacityOperative() float64
}
