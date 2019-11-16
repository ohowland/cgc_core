package cgcintegrationtest

import (
	"log"
	"testing"

	"gotest.tools/assert"
)

type named interface {
	Name() string
}

type power interface {
	KW() float64
}

type capacity interface {
	RealPositive() float64
}

type renewable interface {
	REKW() float64
}

type wt struct {
	kw float64
}

func (a wt) Name() string {
	return "WIND TURBINE"
}

func (a wt) KW() float64 {
	return a.kw
}

func (a wt) REKW() float64 {
	return a.kw
}

type diesel struct {
	kw  float64
	cap float64
}

func (a diesel) Name() string {
	return "DIESEL"
}

func (a diesel) KW() float64 {
	return a.kw
}

func (a diesel) RealPositive() float64 {
	return a.cap
}

func TestPassThrough(t *testing.T) {

	asset1 := wt{10}
	asset2 := diesel{3, 20}

	assets := make([]interface{}, 2)

	assets[0] = asset1
	assets[1] = asset2

	for _, asset := range assets {
		a, ok := asset.(named)
		if ok {
			log.Println(a.Name())
		}

		b, ok := asset.(power)
		if ok {
			log.Println("power:", b.KW())
		}

		c, ok := asset.(capacity)
		if ok {
			log.Println("capacity:", c.RealPositive())
		}

		d, ok := asset.(renewable)
		if ok {
			log.Println("re power:", d.REKW())
		}
	}

	assert.Assert(t, true)
}
