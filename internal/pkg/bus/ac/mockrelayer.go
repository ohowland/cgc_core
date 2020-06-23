package ac

import "math/rand"

type DummyRelay struct {
	hz    float64
	volts float64
}

func (d DummyRelay) Hz() float64 {
	return d.hz
}

func (d DummyRelay) Volts() float64 {
	return d.volts
}

func NewDummyRelay() Relayer {
	return assertedDummyRelay()
}

func randDummyRelayStatus() func() DummyRelay {
	status := DummyRelay{rand.Float64(), rand.Float64()}
	return func() DummyRelay {
		return status
	}
}

var assertedDummyRelay = randDummyRelayStatus()
