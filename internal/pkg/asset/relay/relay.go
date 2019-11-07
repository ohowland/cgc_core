package relay

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(int64, MachineStatus))
}

// Asset is a data structure for a relay
type Asset struct {
	pid    uuid.UUID
	device DeviceController
	status Status
	config Config
}

// PID is a getter for the relay.Asset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.status.setStatus)
}

// WriteControl is unused in relay
func (a Asset) WriteControl() {}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// Status returns the archetypical status for the relay asset.
func (a Asset) Status() Status {
	return a.status
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) Config() Config {
	return a.config
}

// Status is a data structure representing an architypical relay status
type Status struct {
	mux       *sync.Mutex
	timestamp int64
	machine   MachineStatus
}

type MachineStatus struct {
	Hz   float64
	Volt float64
}

// Hz returns relay frequency. Part of the bus.Relayer interface
func (s Status) Hz() float64 {
	return s.machine.Hz
}

// Volt returns relay AC RMS voltage. Part of the bus.Relayer interface
func (s Status) Volt() float64 {
	return s.machine.Volt
}

func (s *Status) setStatus(timestamp int64, ms MachineStatus) {
	if timestamp > s.timestamp { // mux before?
		s.mux.Lock()
		defer s.mux.Unlock()
		s.machine = ms
	}
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	mux     *sync.Mutex
	machine MachineConfig
}

// Config differentiates between two types of configurations, static and dynamic
type MachineConfig struct {
	Name string `json:"Name"`
	Bus  string `json:"Bus"`
}

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	machineConfig := MachineConfig{}
	err := json.Unmarshal(jsonConfig, &machineConfig)
	if err != nil {
		return Asset{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	status := Status{&sync.Mutex{}, 0, MachineStatus{}}
	config := Config{&sync.Mutex{}, machineConfig}

	return Asset{PID, device, status, config}, err

}
