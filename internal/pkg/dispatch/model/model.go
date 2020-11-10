package model

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
)

type Model struct {
	mux   *sync.Mutex
	state State
}

type State struct {
	power    Power
	capacity Capacity
}

type Capacity struct {
	realPositiveCapacity float64
	realNegativeCapacity float64
}

func (c Capacity) RealPositiveCapacity() float64 { return c.realPositiveCapacity }
func (c Capacity) RealNegativeCapacity() float64 { return c.realNegativeCapacity }

type Power struct {
	primaryLoad   float64
	renewableLoad float64
	netLoad       float64
}

func (p Power) PrimaryLoad() float64   { return p.primaryLoad }
func (p Power) RenewableLoad() float64 { return p.renewableLoad }
func (p Power) NetLoad() float64       { return p.netLoad }

func NewModel() (Model, error) {
	return Model{&sync.Mutex{}, State{}}, nil
}

func (m Model) Update(s map[uuid.UUID]dispatch.State) {
	return
}
