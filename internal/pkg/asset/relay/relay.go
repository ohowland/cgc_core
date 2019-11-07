package relay

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for a relay
type Asset struct {
	mux    *sync.Mutex
	pid    uuid.UUID
	device DeviceController
	status Status
	config Config
}

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(Status))
}

// Status is a data structure representing an architypical relay status
type Status struct {
	Timestamp int64
	Hz        float64
	Volt      float64
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Name string `json:"Name"`
	Bus  string `json:"Bus"`
}

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	status := Status{}

	return Asset{&sync.Mutex{}, PID, device, status, config}, err

}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.setStatus)
}

func (a *Asset) setStatus(s Status) {
	if s.Timestamp > a.status.Timestamp { // mux before?
		a.mux.Lock()
		defer a.mux.Unlock()
		a.status = s
	}
}

// WriteControl is unused in relay
func (a Asset) WriteControl() {}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// PID is a getter for the relay.Asset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Hz returns relay frequency. Part of the bus.Relayer interface
func (a Asset) Hz() float64 {
	return a.status.Hz
}

// Volt returns relay AC RMS voltage. Part of the bus.Relayer interface
func (a Asset) Volt() float64 {
	return a.status.Volt
}
