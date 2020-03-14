package asset

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// Controller allows an interface to update and request
// control over Assets (ESS, Grid, PV, etc...)
type Controller interface {
	UpdateStatus()
	RequestControl(uuid.UUID, <-chan msg.Msg) bool
}

// Identifier allows an interface for objects to Identify themselves.
type Identifier interface {
	PID() uuid.UUID
}

//
type Power interface {
	KW() float64
	KVAR() float64
}

//
type Capacity interface {
	RealPositiveCapacity() float64
	RealNegativeCapacity() float64
}
