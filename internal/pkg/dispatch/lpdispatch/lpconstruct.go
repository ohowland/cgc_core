package lpdispatch

import (
	"math"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	opt "github.com/ohowland/cgc_optimize"
)

func (lp *LPDispatch) ConstructUnit(pid uuid.UUID, a interface{})

func asUnit(pid uuid.UUID, a interface{}) opt.Unit {

	var Cp float64 = math.Inf(1)
	var Cn float64 = math.Inf(1)
	var Cc float64 = math.Inf(1)
	c, ok := a.(asset.RealEnergyCost)
	if ok {
		Cp = c.RealPositiveEnergyCost()
		Cn = c.RealNegativeEnergyCost()
		Cc = c.RealCapacityCost()
	}

	var Ce float64 = 0.0

	var XpUb float64 = 0.0
	var XnUb float64 = 0.0
	var XcUb float64 = 0.0
	xc, ok := a.(asset.RealCapacity)
	if ok {
		XpUb = xc.RealPositiveCapacity()
		XnUb = xc.RealNegativeCapacity()

		XcUb = xc.RealPositiveCapacity()
	}

	var XeUb float64 = 0.0
	xe, ok := a.(asset.StoredEnergy)
	if ok {
		XeUb = xe.StoredEnergyCapacity()
	}

	return opt.NewUnit(pid, Cp, Cn, Cc, Ce, XpUb, XnUb, XcUb, XeUb)
}
