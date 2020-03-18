package manualdispatch

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/mock"
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

	status := mock.AssertedStatus()
	msg := msg.New(pid, msg.STATUS, status)

	dispatch.UpdateStatus(msg)

	memberstatus := dispatch.calcStatus.MemberStatus()

	if memberstatus[pid].KW() != status.KW() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid].KW(), status.KW())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid].KW(), status.KW())
	}

	if memberstatus[pid].KVAR() != status.KVAR() {
		t.Errorf("UpdateStatus(): FAILED. %f kVAR != %f kVAR", memberstatus[pid].KVAR(), status.KVAR())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kVAR == %f kVAR", memberstatus[pid].KVAR(), status.KVAR())
	}

	if memberstatus[pid].RealPositiveCapacity() != status.RealPositiveCapacity() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid].RealPositiveCapacity(), status.RealPositiveCapacity())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid].RealPositiveCapacity(), status.RealPositiveCapacity())
	}

	if memberstatus[pid].RealNegativeCapacity() != status.RealNegativeCapacity() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid].RealNegativeCapacity(), status.RealNegativeCapacity())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid].RealNegativeCapacity(), status.RealNegativeCapacity())
	}
}

func TestUpdatePowerMulti(t *testing.T) {
	dispatch, _ := New("./manualdispatch_test_config.json")

	pid1, _ := uuid.NewUUID()
	pid2, _ := uuid.NewUUID()

	status1 := mock.AssertedStatus()
	time.Sleep(100 * time.Millisecond)
	status2 := mock.AssertedStatus()

	msg1 := msg.New(pid1, msg.STATUS, status1)
	dispatch.UpdateStatus(msg1)

	msg2 := msg.New(pid2, msg.STATUS, status2)
	dispatch.UpdateStatus(msg2)

	memberstatus := dispatch.calcStatus.MemberStatus()

	if memberstatus[pid1].KW() != status1.KW() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid1].KW(), status1.KW())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid1].KW(), status1.KW())
	}

	if memberstatus[pid1].KVAR() != status1.KVAR() {
		t.Errorf("UpdateStatus(): FAILED. %f kVAR != %f kVAR", memberstatus[pid1].KVAR(), status1.KVAR())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kVAR == %f kVAR", memberstatus[pid1].KVAR(), status1.KVAR())
	}

	if memberstatus[pid1].RealPositiveCapacity() != status1.RealPositiveCapacity() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid1].RealPositiveCapacity(), status1.RealPositiveCapacity())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid1].RealPositiveCapacity(), status1.RealPositiveCapacity())
	}

	if memberstatus[pid1].RealNegativeCapacity() != status1.RealNegativeCapacity() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid1].RealNegativeCapacity(), status1.RealNegativeCapacity())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid1].RealNegativeCapacity(), status1.RealNegativeCapacity())
	}

	if memberstatus[pid2].KW() != status2.KW() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid2].KW(), status2.KW())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid2].KW(), status2.KW())
	}

	if memberstatus[pid2].KVAR() != status2.KVAR() {
		t.Errorf("UpdateStatus(): FAILED. %f kVAR != %f kVAR", memberstatus[pid2].KVAR(), status2.KVAR())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kVAR == %f kVAR", memberstatus[pid2].KVAR(), status2.KVAR())
	}

	if memberstatus[pid2].RealPositiveCapacity() != status1.RealPositiveCapacity() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid2].RealPositiveCapacity(), status2.RealPositiveCapacity())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid2].RealPositiveCapacity(), status2.RealPositiveCapacity())
	}

	if memberstatus[pid2].RealNegativeCapacity() != status2.RealNegativeCapacity() {
		t.Errorf("UpdateStatus(): FAILED. %f kW != %f kW", memberstatus[pid2].RealNegativeCapacity(), status2.RealNegativeCapacity())
	} else {
		t.Logf("UpdateStatus(): PASSED. %f kW == %f kW", memberstatus[pid2].RealNegativeCapacity(), status2.RealNegativeCapacity())
	}
}

func TestDropAsset(t *testing.T) {
	dispatch, _ := New("./manualdispatch_test_config.json")

	pid1, _ := uuid.NewUUID()
	pid2, _ := uuid.NewUUID()

	status1 := mock.AssertedStatus()
	status2 := mock.AssertedStatus()

	msg1 := msg.New(pid1, msg.STATUS, status1)
	msg2 := msg.New(pid2, msg.STATUS, status2)

	dispatch.UpdateStatus(msg1)
	dispatch.UpdateStatus(msg2)

	dispatch.DropAsset(pid1)

	memberstatus := dispatch.calcStatus.MemberStatus()

	_, ok := memberstatus[pid1]
	assert.Assert(t, !ok)

	if memberstatus[pid2].KW() != status2.KW() {
		t.Errorf("DropAsset(): FAILED. %f != %f", memberstatus[pid2].KW(), status2.KW())
	} else {
		t.Logf("DropAsset(): PASSED. %f == %f", memberstatus[pid2].KW(), status2.KW())
	}
	assert.Assert(t, memberstatus[pid2].KW() == status2.KW())
	assert.Assert(t, memberstatus[pid2].KVAR() == status2.KVAR())
	assert.Assert(t, memberstatus[pid2].RealPositiveCapacity() == status2.RealPositiveCapacity())
	assert.Assert(t, memberstatus[pid2].RealNegativeCapacity() == status2.RealNegativeCapacity())
}
