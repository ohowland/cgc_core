package manualdispatch

import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"gotest.tools/assert"
	"testing"
)

func TestNew(t *testing.T) {
	_, err := New("")
	if err != nil {
		t.Fatal(err)
	}
}

type MockStatus struct {
	Real    float64
	React   float64
	Realpos float64
	Realneg float64
}

func (m MockStatus) RealPositiveCapacity() float64 {
	return m.Realpos
}

func (m MockStatus) RealNegativeCapacity() float64 {
	return m.Realneg
}

func (m MockStatus) KW() float64 {
	return m.Real
}
func (m MockStatus) KVAR() float64 {
	return m.React
}

func TestUpdateStatusSingle(t *testing.T) {
	dispatch, _ := New("")

	pid, _ := uuid.NewUUID()

	status := MockStatus{10, 20, 30, 40}
	msg := asset.NewMsg(pid, status)

	dispatch.UpdateStatus(msg)

	memberstatus := dispatch.calcStatus.MemberStatus()

	assert.Assert(t, memberstatus[pid].KW() == status.KW())
	assert.Assert(t, memberstatus[pid].KVAR() == status.KVAR())
	assert.Assert(t, memberstatus[pid].RealPositiveCapacity() ==
		status.RealPositiveCapacity())
	assert.Assert(t, memberstatus[pid].RealNegativeCapacity() ==
		status.RealNegativeCapacity())
}

func TestUpdatePowerMulti(t *testing.T) {
	dispatch, _ := New("")

	pid1, _ := uuid.NewUUID()
	pid2, _ := uuid.NewUUID()

	status1 := MockStatus{11, 22, 33, 44}
	status2 := MockStatus{55, 66, 77, 88}
	msg := asset.NewMsg(pid1, status1)
	dispatch.UpdateStatus(msg)

	msg = asset.NewMsg(pid2, status2)
	dispatch.UpdateStatus(msg)

	memberstatus := dispatch.calcStatus.MemberStatus()

	assert.Assert(t, memberstatus[pid1].KW() == status1.KW())
	assert.Assert(t, memberstatus[pid1].KVAR() == status1.KVAR())
	assert.Assert(t, memberstatus[pid1].RealPositiveCapacity() ==
		status1.RealPositiveCapacity())
	assert.Assert(t, memberstatus[pid1].RealNegativeCapacity() ==
		status1.RealNegativeCapacity())

	assert.Assert(t, memberstatus[pid2].KW() == status2.KW())
	assert.Assert(t, memberstatus[pid2].KVAR() == status2.KVAR())
	assert.Assert(t, memberstatus[pid2].RealPositiveCapacity() ==
		status2.RealPositiveCapacity())
	assert.Assert(t, memberstatus[pid2].RealNegativeCapacity() ==
		status2.RealNegativeCapacity())
}

func TestDropAsset(t *testing.T) {
	dispatch, _ := New("")

	pid1, _ := uuid.NewUUID()
	pid2, _ := uuid.NewUUID()

	status1 := MockStatus{11, 22, 33, 44}
	status2 := MockStatus{55, 66, 77, 88}

	msg1 := asset.NewMsg(pid1, status1)
	msg2 := asset.NewMsg(pid2, status2)
	dispatch.UpdateStatus(msg1)
	dispatch.UpdateStatus(msg2)

	dispatch.DropAsset(pid1)

	memberstatus := dispatch.calcStatus.MemberStatus()

	_, ok := memberstatus[pid1]
	assert.Assert(t, !ok)

	assert.Assert(t, memberstatus[pid2].KW() == status2.KW())
	assert.Assert(t, memberstatus[pid2].KVAR() == status2.KVAR())
	assert.Assert(t, memberstatus[pid2].RealPositiveCapacity() == status2.RealPositiveCapacity())
	assert.Assert(t, memberstatus[pid2].RealNegativeCapacity() == status2.RealNegativeCapacity())
}
