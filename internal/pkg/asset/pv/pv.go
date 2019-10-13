package pv

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Asset is a datastructure for an PV Asset
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

// Status is a data structure representing an architypical PV status
type Status struct {
	Timestamp int64
	KW        float64
	KVAR      float64
	Hz        float64
	Volt      float64
	Online    bool
}

// Control is a data structure representing an architypical PV control
type Control struct {
	Run     bool
	Enable  bool
	KWLimit float64
	KVAR    float64
}

type Config struct {
	Name      string  `json:"Name"`
	Bus       string  `json:"Bus"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
}

// PID is a getter for the unique identifier field
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Status is a getter for the pv.Asset status field
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

// WriteControl requests a physical device write of the data held in the PV control field.
func (a Asset) WriteControl() {
	go a.device.WriteDeviceControl(a.control)
}

// SetControl is a setter for the PV control field
func (a *Asset) SetControl(c Control) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control = c
}

// New returns a configured pv.Asset
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

func (a Asset) Name() string {
	return a.config.Name
}

func (a Asset) KW() float64 {
	return a.Status().KW
}

func (a Asset) KVAR() float64 {
	return a.Status().KVAR
}

func (a *Asset) KWCmd(kw float64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.KWLimit = kw
}

func (a *Asset) KVARCmd(kvar float64) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.KVAR = kvar
}

func (a *Asset) RunCmd(run bool) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.control.Run = run
}
