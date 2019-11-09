package virtualacbus

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
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

func newVirtualBus() *VirtualACBus {
	configpath := "../acbus_test_config.json"
	bus, err := New(configpath)
	if err != nil {
		panic(err)
	}
	return bus.(*acbus.ACBus).Relayer().(*VirtualACBus)
}

func TestNewVirtualACBus(t *testing.T) {
	configpath := "../acbus_test_config.json"
	bus, err := New(configpath)
	if err != nil {
		t.Fatal(err)
	}
	acbus := bus.(acbus.ACBus)
	assert.Assert(t, acbus.Name() == "TEST_Virtual Bus")
}

func TestCalcSwingLoad(t *testing.T) {
	bus := newVirtualBus()

	asset1 := NewDummyAsset()
	asset2 := NewDummyAsset()

	bus.assetObserver <- asset1.asSource()
	bus.assetObserver <- asset2.asSource()

	gridformer := bus.gridformer

	assertKwSum := asset1.status.kW + asset2.status.kW
	assertKvarSum := asset1.status.kVAR + asset2.status.kVAR

	assert.Assert(t, gridformer.KW == assertKwSum)
	assert.Assert(t, gridformer.KVAR == assertKvarSum)

	close(bus.assetObserver)
}

func TestCalcSwingLoadChange(t *testing.T) {
	bus := newVirtualBus()

	gridfollowingAsset1 := NewDummyAsset()
	gridfollowingAsset2 := NewDummyAsset()
	gridformingAsset := NewDummyAsset()

	gridformingAsset.status.gridforming = true

	bus.assetObserver <- gridfollowingAsset1.asSource()
	bus.assetObserver <- gridformingAsset.asSource()
	bus.assetObserver <- gridfollowingAsset2.asSource()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 1; i <= 5; i++ {
		gridfollowingAsset1.status.kW = r.Float64() * 1000
		gridfollowingAsset2.status.kW = r.Float64() * 1000

		bus.assetObserver <- gridfollowingAsset2.asSource()
		bus.assetObserver <- gridformingAsset.asSource()
		bus.assetObserver <- gridfollowingAsset1.asSource()
	}

	time.Sleep(time.Duration(200) * time.Millisecond)

	gridformer := bus.gridformer
	assertKwSum := gridfollowingAsset1.status.kW + gridfollowingAsset2.status.kW
	assertKvarSum := gridfollowingAsset1.status.kVAR + gridfollowingAsset2.status.kVAR
	assert.Assert(t, gridformer.KW == assertKwSum)
	assert.Assert(t, gridformer.KVAR == assertKvarSum)

	close(bus.assetObserver)
}
