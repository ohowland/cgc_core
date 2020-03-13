package manualdispatch

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/msg"
	"gotest.tools/assert"
)

func TestNew(t *testing.T) {
	_, err := New("./manualdispatch_test_config.json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateStatusSingle(t *testing.T) {
	dispatch, _ := New("./manualdispatch_test_config.json")

	pid, _ := uuid.NewUUID()

	status := asset.DummyStatus{}
	msg := msg.New(pid, status)

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
	dispatch, _ := New("./manualdispatch_test_config.json")

	pid1, _ := uuid.NewUUID()
	pid2, _ := uuid.NewUUID()

	status1 := MockAsset{MockStatus{"ESS", pid1, 10, 20, 30, 40}}
	status2 := MockAsset{MockStatus{"Grid", pid2, 40, 50, 60, 70}}
	msg1 := msg.New(pid1, status1)
	dispatch.UpdateStatus(msg1)

	msg2 := msg.New(pid2, status2)
	dispatch.UpdateStatus(msg2)

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
	dispatch, _ := New("./manualdispatch_test_config.json")

	pid1, _ := uuid.NewUUID()
	pid2, _ := uuid.NewUUID()

	status1 := MockAsset{MockStatus{"ESS", pid1, 11, 22, 33, 44}}
	status2 := MockAsset{MockStatus{"Grid", pid2, 55, 66, 77, 88}}

	msg1 := msg.New(pid1, status1)
	msg2 := msg.New(pid2, status2)
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
