package virtualpv

import (
	"fmt"
	"math"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func newArrayDeg(tilt float64, azmuth float64) Array {
	degToRad := math.Pi / 180
	return Array{
		tiltAngle:    tilt * degToRad,
		azimuthAngle: azmuth * degToRad}
}

func newLocation(lat float64, elevFt float64) Location {
	degToRad := math.Pi / 180
	ftToKm := 0.0003048

	return Location{
		latitude:  lat * degToRad,
		elevation: elevFt * ftToKm,
	}
}

func TestSunrise(t *testing.T) {
	l := newLocation(42, 5000)
	t1 := time.Now()
	t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 12, 0, 0, 0, t1.Location())

	sr := sunrise(l, t2)

	assert.Assert(t, sr.Before(t2))
}

func TestSunset(t *testing.T) {
	l := newLocation(42, 5000)
	t1 := time.Now()
	t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 12, 0, 0, 0, t1.Location())

	ss := sunset(l, t2)

	assert.Assert(t, ss.After(t2))
}

func TestSunsetAfterSunrise(t *testing.T) {
	l := newLocation(42, 5000)
	t1 := time.Now()
	t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 12, 0, 0, 0, t1.Location())

	sr := sunrise(l, t2)
	ss := sunset(l, t2)

	assert.Assert(t, ss.After(sr))
}

func TestTotalIrradiance(t *testing.T) {
	a := newArrayDeg(32, 0)
	l := newLocation(42, 5000)
	t1 := time.Now()
	t2 := time.Date(t1.Year(), t1.Month(), t1.Day(), 12, 0, 0, 0, t1.Location())
	r := TotalIrradiance(a, l, t2)

	fmt.Printf("irradiance: %v\n", r)
	assert.Assert(t, r > 0)
}
