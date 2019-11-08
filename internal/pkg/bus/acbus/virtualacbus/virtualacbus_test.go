package virtualacbus

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

type DummyAsset struct {
	pid    uuid.UUID
	status Status
}

type Status struct {
	kW          float64
	kVAR        float64
	hz          float64
	volt        float64
	gridforming bool
}

func NewDummyAsset() DummyAsset {
	pid, _ := uuid.NewUUID()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return DummyAsset{
		pid: pid,
		status: Status{
			kW:          r.Float64() * 10000,
			kVAR:        r.Float64() * 10000,
			hz:          60,
			volt:        480,
			gridforming: false,
		},
	}
}

func (a DummyAsset) asSource() Source {
	return Source{
		PID:         a.pid,
		Hz:          a.status.hz,
		Volt:        a.status.volt,
		KW:          a.status.kW,
		KVAR:        a.status.kVAR,
		Gridforming: a.status.gridforming,
	}
}

func TestNewVirtualACBus(t *testing.T) {

	bus := New()
	time.Sleep(time.Duration(200) * time.Millisecond)
	close(bus.observer)
	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestCalcSwingLoad(t *testing.T) {
	bus := New()

	asset1 := NewDummyAsset()
	time.Sleep(time.Duration(1) * time.Millisecond)
	asset2 := NewDummyAsset()

	bus.observer <- asset1.asSource()
	bus.observer <- asset2.asSource()

	time.Sleep(time.Duration(200) * time.Millisecond)

	gridformer := bus.Gridformer()

	assertKwSum := asset1.status.kW + asset2.status.kW
	assertKvarSum := asset1.status.kVAR + asset2.status.kVAR

	assert.Assert(t, gridformer.KW == assertKwSum)
	assert.Assert(t, gridformer.KVAR == assertKvarSum)

	close(bus.observer)
}

func TestCalcSwingLoadChange(t *testing.T) {
	bus := New()

	gridfollowingAsset1 := NewDummyAsset()
	time.Sleep(time.Duration(1) * time.Millisecond)
	gridfollowingAsset2 := NewDummyAsset()
	time.Sleep(time.Duration(1) * time.Millisecond)
	gridformingAsset := NewDummyAsset()

	gridformingAsset.status.gridforming = true

	bus.observer <- gridfollowingAsset1.asSource()
	bus.observer <- gridformingAsset.asSource()
	bus.observer <- gridfollowingAsset2.asSource()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 1; i <= 5; i++ {
		gridfollowingAsset1.status.kW = r.Float64() * 1000
		time.Sleep(time.Duration(10) * time.Millisecond)
		gridfollowingAsset2.status.kW = r.Float64() * 1000

		bus.observer <- gridfollowingAsset2.asSource()
		bus.observer <- gridformingAsset.asSource()
		bus.observer <- gridfollowingAsset1.asSource()
	}

	time.Sleep(time.Duration(200) * time.Millisecond)

	gridformer := bus.Gridformer()
	assertKwSum := gridfollowingAsset1.status.kW + gridfollowingAsset2.status.kW
	assertKvarSum := gridfollowingAsset1.status.kVAR + gridfollowingAsset2.status.kVAR
	assert.Assert(t, gridformer.KW == assertKwSum)
	assert.Assert(t, gridformer.KVAR == assertKvarSum)

	close(bus.observer)
}
