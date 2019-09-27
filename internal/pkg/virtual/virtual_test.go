package virtual

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

type DummyAsset struct {
	id     uuid.UUID
	status Status
}

type Status struct {
	KW   float64
	KVAR float64
}

func NewDummyAsset() DummyAsset {
	id, _ := uuid.NewUUID()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return DummyAsset{
		id: id,
		status: Status{
			KW:   r.Float64() * 10000,
			KVAR: r.Float64() * 10000,
		},
	}
}

func (a DummyAsset) load() SourceLoad {
	return SourceLoad{
		ID: a.id,
		Load: Load{
			KW:   a.status.KW,
			KVAR: a.status.KVAR,
		},
	}
}

func TestNewVirtualSystemModel(t *testing.T) {

	vsm := NewVirtualSystemModel()

	go vsm.RunVirtualSystem()
	time.Sleep(time.Duration(200) * time.Millisecond)
	vsm.StopVirtualSystem()
	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestCalcSwingLoad(t *testing.T) {
	vsm := NewVirtualSystemModel()
	go vsm.RunVirtualSystem()

	asset1 := NewDummyAsset()
	time.Sleep(time.Duration(1) * time.Millisecond)
	asset2 := NewDummyAsset()

	vsm.ReportLoad <- asset1.load()
	vsm.ReportLoad <- asset2.load()

	time.Sleep(time.Duration(200) * time.Millisecond)

	swingload := <-vsm.SwingLoad
	swingload = <-vsm.SwingLoad

	assertKwSum := asset1.status.KW + asset2.status.KW
	assertKvarSum := asset1.status.KVAR + asset2.status.KVAR

	assert.Assert(t, swingload.KW == assertKwSum)
	assert.Assert(t, swingload.KVAR == assertKvarSum)

	vsm.StopVirtualSystem()
}

func TestCalcSwingLoadChange(t *testing.T) {
	vsm := NewVirtualSystemModel()
	go vsm.RunVirtualSystem()

	asset1 := NewDummyAsset()
	time.Sleep(time.Duration(1) * time.Millisecond)
	asset2 := NewDummyAsset()

	vsm.ReportLoad <- asset1.load()
	vsm.ReportLoad <- asset2.load()

	time.Sleep(time.Duration(200) * time.Millisecond)

	swingload := <-vsm.SwingLoad
	swingload = <-vsm.SwingLoad

	asset1.status.KW = 1
	vsm.ReportLoad <- asset1.load()

	time.Sleep(time.Duration(200) * time.Millisecond)

	swingload = <-vsm.SwingLoad
	swingload = <-vsm.SwingLoad

	assertKwSum := asset1.status.KW + asset2.status.KW
	assertKvarSum := asset1.status.KVAR + asset2.status.KVAR

	assert.Assert(t, swingload.KW == assertKwSum)
	assert.Assert(t, swingload.KVAR == assertKvarSum)

	vsm.StopVirtualSystem()
}
