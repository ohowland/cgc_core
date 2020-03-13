package asset

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// Controller is the interface to update Asset Status
type Controller interface {
	UpdateStatus()
	RequestControl(uuid.UUID, <-chan msg.Msg) bool
}

type Identifier interface {
	PID() uuid.UUID
	Name() string
}

type Power interface {
	KW() float64
	KVAR() float64
}

type Capacity interface {
	RealPositiveCapacity() float64
	RealNegativeCapacity() float64
}
