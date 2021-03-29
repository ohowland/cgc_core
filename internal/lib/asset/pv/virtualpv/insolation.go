package virtualpv

import (
	"math"
	"time"
)

// Array specifies all angles in radians
type Array struct {
	tiltAngle    float64
	azimuthAngle float64
}

// Location all angles in radians, elevation in km
type Location struct {
	latitude  float64
	elevation float64
}

// Radiation in w/m^2)
type Radiation struct {
	direct  float64
	diffuse float64
}

func TotalIrradiance(a Array, l Location, t time.Time) float64 {
	rad := intensity(l, t)
	angle := incidentAngle(l, a, t)

	if angle > math.Pi/2 {
		return rad.diffuse
	}

	return rad.direct*math.Cos(angle) + rad.diffuse
}

func intensity(l Location, t time.Time) Radiation {
	if isDaytime(l, t) {
		x1 := math.Pow(0.7, math.Pow(airMass(l, t), 0.678))
		x2 := l.elevation * 0.14

		y1 := (x1*(1-x2) + x2) * 1353

		return Radiation{y1, y1 * 0.1}
	}
	return Radiation{0, 0}
}

func incidentAngle(l Location, a Array, t time.Time) float64 {
	x1 := math.Cos(hourAngle(t))

	d := declinationAngle(t)
	y1 := math.Cos(d)
	y2 := math.Sin(d)

	z1 := math.Cos(l.latitude - a.tiltAngle)
	z2 := math.Sin(l.latitude - a.tiltAngle)

	angle := math.Acos(x1*y1*z1 + y2*z2)
	return angle
}

func airMass(l Location, t time.Time) float64 {
	return 1 / (math.Cos((math.Pi / 2) - elevationAngle(l, t)))
}

func elevationAngle(l Location, t time.Time) float64 {
	d := declinationAngle(t)
	x1 := math.Sin(d)
	y1 := math.Sin(l.latitude)

	z1 := x1 * y1

	x2 := math.Cos(d)
	y2 := math.Cos(l.latitude)

	z2 := x2 * y2 * math.Cos(hourAngle(t))

	elevAngle := math.Asin(z1 + z2)

	if elevAngle < 0 {
		return 0
	}

	return elevAngle
}

func hourAngle(t time.Time) float64 {

	hourOfDay := float64(t.Hour()*3600+t.Minute()*60+t.Second()) / 3600
	return (hourOfDay - 12) * 15 * (math.Pi / 180)
}

func isDaytime(l Location, t time.Time) bool {
	if t.After(sunrise(l, t)) && t.Before(sunset(l, t)) {
		return true
	}
	return false
}

func sunset(l Location, t time.Time) time.Time {
	d := declinationAngle((t))
	x1 := math.Sin(d)
	x2 := math.Sin(l.latitude)

	y1 := math.Cos(d)
	y2 := math.Cos(l.latitude)

	z1 := -1 * (x1 * x2 / y1 * y2)
	z2 := (math.Acos(z1) * (180 / math.Pi)) / 15

	setHr, fracHr := math.Modf(12 + z2)
	setMin, fracMin := math.Modf(fracHr * 60)
	setSec, _ := math.Modf(fracMin * 60)

	tSunset := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		int(setHr),
		int(setMin),
		int(setSec),
		0,
		t.Location(),
	)

	return tSunset
}

func sunrise(l Location, t time.Time) time.Time {
	d := declinationAngle((t))
	x1 := math.Sin(d)
	x2 := math.Sin(l.latitude)

	y1 := math.Cos(d)
	y2 := math.Cos(l.latitude)

	z1 := -1 * (x1 * x2 / y1 * y2)
	z2 := (math.Acos(z1) * (180 / math.Pi)) / 15

	riseHr, fracHr := math.Modf(12 - z2)
	riseMin, fracMin := math.Modf(fracHr * 60)
	riseSec, _ := math.Modf(fracMin * 60)

	tSunrise := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		int(riseHr),
		int(riseMin),
		int(riseSec),
		0,
		t.Location(),
	)

	return tSunrise
}

func declinationAngle(t time.Time) float64 {
	x1 := math.Sin(((float64(t.YearDay()) - 81) * 2 * math.Pi) / 365.25)
	x2 := math.Sin(0.40928)

	return math.Asin(x1 * x2)
}
