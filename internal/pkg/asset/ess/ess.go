package ess

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus() (MachineStatus, error)
	WriteDeviceControl(MachineControl) error
}

// Asset is a data structure for an ESS Asset
type Asset struct {
	mux          *sync.Mutex
	pid          uuid.UUID
	device       DeviceController
	publisher    *msg.PubSub
	controlOwner uuid.UUID
	supervisory  SupervisoryControl
	config       Config
}

// PID is a getter for the asset PID
func (a Asset) PID() uuid.UUID {
	return a.pid
}

// Name is a getter for the asset Name
func (a Asset) Name() string {
	return a.config.machine.Name
}

// BusName is a getter for the asset's connected Bus
func (a Asset) BusName() string {
	return a.config.machine.BusName
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// Subscribe returns a channel on which the specified topic is broadcast
func (a Asset) Subscribe(pid uuid.UUID, topic msg.Topic) <-chan msg.Msg {
	ch := a.publisher.Subscribe(pid, topic)
	return ch
}

// Unsubscribe pid from all topic broadcasts
func (a Asset) Unsubscribe(pid uuid.UUID) {
	a.publisher.Unsubscribe(pid)
}

// RequestControl connects the asset control to the read only channel parameter.
func (a *Asset) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) bool {
	a.mux.Lock()
	defer a.mux.Unlock()
	// TODO: previous owner needs to stop. how to enforce?
	a.controlOwner = pid
	go a.controlHandler(ch)

	return true
}

// UpdateStatus requests a physical device read, then broadcasts results
func (a Asset) UpdateStatus() {
	machineStatus, err := a.device.ReadDeviceStatus()
	if err != nil {
		// Read Error Handler Path
		return
	}
	status := transform(machineStatus)
	a.publisher.Publish(msg.Status, status)
}

// UpdateConfig requests component broadcast current configuration
func (a Asset) UpdateConfig() {
	a.publisher.Publish(msg.Config, a.Config())
}

func transform(machineStatus MachineStatus) Status {
	return Status{
		CalculatedStatus{},
		machineStatus,
	}
}

func (a *Asset) controlHandler(ch <-chan msg.Msg) {
loop:
	for {
		data, ok := <-ch
		if !ok {
			log.Println("ESS controlHandler() stopping")
			break loop
		}
		control, ok := data.Payload().(MachineControl)
		if !ok {
			log.Println("ESS controlHandler() bad type assertion")
			continue
		}
		err := a.device.WriteDeviceControl(control)
		if err != nil {
			// TODO: Write Error Handler Path
			log.Println("ESS controlHandler():", err)
		}
	}
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) Config() MachineConfig {
	return a.config.machine
}

// Status wraps MachineStatus with mutex and state metadata
type Status struct {
	Calc    CalculatedStatus `json:"CalculatedStatus"`
	Machine MachineStatus    `json:"MachineStatus"`
}

// CalculatedStatus is a data structure representing asset state information
// that is calculated from data read into the archetype ess.
type CalculatedStatus struct{}

// MachineStatus is a data structure representing an architypical ESS status
type MachineStatus struct {
	KW                   float64 `json:"KW"`
	KVAR                 float64 `json:"KVAR"`
	Hz                   float64 `json:"Hz"`
	Volt                 float64 `json:"Volt"`
	RealPositiveCapacity float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity float64 `json:"RealNegativeCapacity"`
	SOC                  float64 `json:"SOC"`
	Gridforming          bool    `json:"Gridforming"`
	Online               bool    `json:"Online"`
}

// KW returns the asset's measured real power
func (s Status) KW() float64 {
	return s.Machine.KW
}

// KVAR returns the asset's measured reactive power
func (s Status) KVAR() float64 {
	return s.Machine.KVAR
}

// RealPositiveCapacity returns the asset's operative real positive capacity
func (s Status) RealPositiveCapacity() float64 {
	return s.Machine.RealPositiveCapacity
}

// RealNegativeCapacity returns the asset's operative real negative capacity
func (s Status) RealNegativeCapacity() float64 {
	return s.Machine.RealNegativeCapacity
}

// MachineControl defines the hardware control interface for the ESS Asset
type MachineControl struct {
	Run      bool
	KW       float64
	KVAR     float64
	Gridform bool
}

// SupervisoryControl defines the software control interface for the ESS Asset
type SupervisoryControl struct {
	mux    *sync.Mutex
	enable bool
}

// Config wraps MachineConfig with mutex a mutex and hides the internal state.
type Config struct {
	mux     *sync.Mutex
	machine MachineConfig
}

// MachineConfig holds the ESS asset configuration parameters
type MachineConfig struct {
	Name      string  `json:"Name"`
	BusName   string  `json:"BusName"`
	RatedKW   float64 `json:"RatedKW"`
	RatedKVAR float64 `json:"RatedKVAR"`
	RatedKWH  float64 `json:"RatedKWH"`
}

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	machineConfig := MachineConfig{}
	err := json.Unmarshal(jsonConfig, &machineConfig)
	if err != nil {
		return Asset{}, err
	}

	pid, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	publisher := msg.NewPublisher(pid)

	var controlOwner uuid.UUID

	supervisory := SupervisoryControl{&sync.Mutex{}, false}
	config := Config{&sync.Mutex{}, machineConfig}

	return Asset{&sync.Mutex{}, pid, device, publisher, controlOwner, supervisory, config}, err
}
