package relay

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for a relay
type Asset struct {
	mux     *sync.Mutex
	pid     uuid.UUID
	device  DeviceController
	status  Status
	control Control
	config  Config
}

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(Status))
	WriteDeviceControl(Control)
}

// Status is a data structure representing an architypical relay status
type Status struct {
	Timestamp int64
	Hz        float64
	Volt      float64
}

// Control is a data structure representing an architypical relay control
type Control struct{}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Static StaticConfig `json:"StaticConfig"`
}

// StaticConfig is a data structure representing an architypical fixed relay configuration
type StaticConfig struct {
	Name string `json:"Name"`
}

// PID is a getter for the relay.Asset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Status is a getter for the relay.Asset status field
func (a Asset) Status() Status {
	return a.status
}

// UpdateStatus requests a physical device read and updates the ess.Asset status field
func (a *Asset) UpdateStatus() {
	go a.device.ReadDeviceStatus(a.setStatus)
}

func (a *Asset) setStatus(s Status) {
	if a.filterTimestamp(s.Timestamp) {
		a.mux.Lock()
		defer a.mux.Unlock()
		a.status = s
	}
}

// WriteControl requests a physical device write of the data held in the GridAsset control field.
func (a Asset) WriteControl() {
	go a.device.WriteDeviceControl(a.control)
}

// SetControl is a setter for the ess.Asset control field
func (a *Asset) SetControl(c Control) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control = c
}

// StaticConfig is a getter for the relay.Asset static config field
func (a Asset) StaticConfig() StaticConfig {
	return a.config.Static
}

// Hz returns relay frequency. Part of the Relayer interface
func (a Asset) Hz() float64 {
	return a.status.Hz
}

// Volt returns relay AC RMS voltage. Part of the Relayer interface
func (a Asset) Volt() float64 {
	return a.status.Volt
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
	control := Control{}

	return Asset{&sync.Mutex{}, PID, device, status, control, config}, err

}

func (a *Asset) filterTimestamp(timestamp int64) bool {
	return timestamp > a.status.Timestamp
}
