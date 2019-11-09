package acbus

import (
	"testing"

	"github.com/google/uuid"
)

type DummyAsset struct {
	pid uuid.UUID
}

type DummyStatus struct {
	kW                   float64
	kVAR                 float64
	hz                   float64
	volt                 float64
	realPositiveCapacity float64
	realNegativeCapacity float64
	gridforming          bool
}

func (a DummyAsset) PID() uuid.UUID {
	return a.pid
}

func (s DummyStatus) KW() float64 {
	return s.kW
}

func (s DummyStatus) KVAR() float64 {
	return s.kVAR
}

func (s DummyStatus) RealPositiveCapacity() float64 {
	return s.realPositiveCapacity
}

func (s DummyStatus) RealNegativeCapacity() float64 {
	return s.realNegativeCapacity
}

func NewDummyAsset() DummyAsset {
	pid, _ := uuid.NewUUID()
	return DummyAsset{
		pid: pid,
	}
}

func TestNewAcBus(t *testing.T) {
	configPath := "./acbus_test_config.json"
}

func TestAddMember(t *testing.T) {}

func TestRemoveMember(t *testing.T) {}

func TestProcess(t *testing.T) {}

func TestAggregateCapacity(t *testing.T) {}

func TestUpdateRelayer(t *testing.T) {}
