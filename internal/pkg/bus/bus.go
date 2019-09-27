package asset

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type Relayer interface {
	Hz() float64
	Volt() float64
}

type Bus struct {
	id      uuid.UUID
	relay   Relayer
	members map[uuid.UUID]asset.Asset
	status  Status
}

type Status struct {
	Hz        float64 `json:"Hz"`
	Volts     float64 `json:"Volts"`
	Energized bool    `json:"Energized"`
}

/*
TODO:

1. the bus object constructs a bus graph.
2. the virtual system model class should request members from the bus,
poll those members for load information then report the swing load to the
gridformer on the bus.
3.
*/
