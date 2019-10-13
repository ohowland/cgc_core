package feeder

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a data structure for an Feeder Asset
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

// Status is a data structure representing an architypical Feeder status
type Status struct {
	Timestamp int64
	KW        float64
	KVAR      float64
	Hz        float64
	Volt      float64
	Online    bool
}

// Control is a data structure representing an architypical Feeder control
type Control struct {
	CloseFeeder bool
	Enable      bool
}

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
}

// StaticConfig is a data structure representing an architypical fixed feeder configuration
type StaticConfig struct { // Should this get transfered over to the specific class?

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

// SetControl is a setter for the ess.Asset control field
func (a *Asset) SetControl(c Control) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control = c
}

// WriteControl requests a physical device write of the data held in the GridAsset control field.
func (a Asset) WriteControl() {
	go a.device.WriteDeviceControl(a.control)
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

func (a Asset) Name() string {
	return a.config.Name
}

func (a Asset) KW() float64 {
	return a.Status().KW
}

func (a Asset) KVAR() float64 {
	return a.Status().KVAR
}

func (a *Asset) RunCmd(run bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.CloseFeeder = run
}
