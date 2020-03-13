package manualdispatch

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
	"gotest.tools/assert"
)

type MockAsset struct {
	Status MockStatus `json:"Status"`
}

type MockStatus struct {
	Name                 string    `json:"Name"`
	PID                  uuid.UUID `json:"PID"`
	KW                   float64   `json:"KW"`
	KVAR                 float64   `json:"KVAR"`
	RealPositiveCapacity float64   `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64   `json:"RealNegativeCapacity"`
}

func (s MockAsset) KW() float64 {
	return s.Status.KW
}
func (s MockAsset) KVAR() float64 {
	return s.Status.KVAR
}
func (s MockAsset) RealPositiveCapacity() float64 {
	return s.Status.RealPositiveCapacity
}
func (s MockAsset) RealNegativeCapacity() float64 {
	return s.Status.RealNegativeCapacity
}

func TestNew(t *testing.T) {
	_, err := New("./manualdispatch_test_config.json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateStatusSingle(t *testing.T) {
	dispatch, _ := New("./manualdispatch_test_config.json")

	pid, _ := uuid.NewUUID()

	status := MockAsset{MockStatus{"ESS", pid, 10, 20, 30, 40}}
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
	msg := msg.New(pid1, status1)
	dispatch.UpdateStatus(msg)

	msg = msg.New(pid2, status2)
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
