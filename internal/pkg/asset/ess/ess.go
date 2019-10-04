package ess

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for an ESS Asset
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

// Status is a data structure representing an architypical ESS status
type Status struct {
	Timestamp            int64
	KW                   float64
	KVAR                 float64
	Hz                   float64
	Volt                 float64
	SOC                  float64
	PositiveRealCapacity float64
	NegativeRealCapacity float64
	Gridforming          bool
	Online               bool
}

// Control is a data structure representing an architypical ESS control
type Control struct {
	Run      bool
	Enable   bool
	KW       float64
	KVAR     float64
	GridForm bool
}

type Config struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	KWRated   float64 `json:"KWRated"`
	KVARRated float64 `json:"KVARRated"`
	KWHRated  float64 `json:"KWHRated"`
}

// PID is a getter for the ess.Asset status field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Status is a getter for the ess.Asset status field
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

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	PID, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	config := Config{}
	err = json.Unmarshal(jsonConfig, &config)
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
