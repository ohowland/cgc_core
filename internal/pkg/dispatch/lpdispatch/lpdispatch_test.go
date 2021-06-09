package lpdispatch

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset/mockasset"
	opt "github.com/ohowland/cgc_optimize"
	"github.com/ohowland/highs"
	"gotest.tools/v3/assert"
)

func TestSolverCall(t *testing.T) {
	pid, _ := uuid.NewUUID()
	u1 := opt.NewUnit(pid, 1, 1, 1, 0, 5, 5, 5, 0)
	u1.NewConstraint(opt.UnitCapacityConstraints(&u1)...)
	fmt.Printf("%+v\n", u1)
	u2 := opt.NewUnit(pid, 2, 2, 2, 0, 5, 5, 5, 0)
	u2.NewConstraint(opt.UnitCapacityConstraints(&u2)...)

	g1 := opt.NewGroup(u1, u2)
	g1.NewConstraint(opt.NetLoadConstraint(&g1, 9))

	s, err := highs.New(
		g1.CostCoefficients(),
		g1.Bounds(),
		g1.Constraints(),
		[]int{})

	assert.NilError(t, err)

	s.SetObjectiveSense(highs.Minimize)
	s.RunSolver()
	assert.NilError(t, err)
	fmt.Println(s.PrimalColumnSolution())

}

func TestBuildUnitandSolve(t *testing.T) {
	pid, _ := uuid.NewUUID()

	a1 := mockasset.AssertedStatus()
	u1 := BuildUnit(pid, a1)
	u2 := BuildUnit(pid, a1)

	g1 := opt.NewGroup(u1, u2)
	g1.NewConstraint(opt.NetLoadConstraint(&g1, 0.1))

	s, err := highs.New(
		g1.CostCoefficients(),
		g1.Bounds(),
		g1.Constraints(),
		[]int{})

	assert.NilError(t, err)

	s.SetObjectiveSense(highs.Minimize)
	s.RunSolver()

	fmt.Println(s.PrimalColumnSolution())
}
