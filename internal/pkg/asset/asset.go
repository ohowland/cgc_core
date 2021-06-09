package asset

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// Asset interface, anything with a name.
type Asset interface {
	Controller
	Config
	msg.Publisher
}

// Controller allows an interface to update and request
// control over Assets (ESS, Grid, PV, etc...)
type Controller interface {
	UpdateStatus()
	UpdateConfig()
	RequestControl(uuid.UUID, <-chan msg.Msg) error
	Shutdown(*sync.WaitGroup) error
}

type Config interface {
	PID() uuid.UUID
	Name() string
	BusName() string
}

//
type RealPower interface {
	KW() float64
}

type RealEnergyCost interface {
	RealPositiveEnergyCost() float64
	RealNegativeEnergyCost() float64
	RealCapacityCost() float64
}

type ReactivePower interface {
	KVAR() float64
}

type Voltage interface {
	Volts() float64
}

type Frequency interface {
	Hz() float64
}

type Gridforming interface {
	Gridforming() bool
}

//
type RealCapacity interface {
	RealPositiveCapacity() float64
	RealNegativeCapacity() float64
}

type StoredEnergy interface {
	StoredEnergy() float64
	StoredEnergyCapacity() float64
}
