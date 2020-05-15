package relay

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus(func(int64, MachineStatus))
}

// Asset is a data structure for a relay
type Asset struct {
	pid       uuid.UUID
	device    DeviceController
	publisher *msg.PubSub
	config    Config
}

// PID is a getter for the relay.Asset status field
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

// UpdateStatus requests a physical device read, then updates MachineStatus field.
func (a Asset) UpdateStatus() {
	machineStatus, err := a.device.ReadDeviceStatus()
	if err != nil {
		// Read Error Handler Path
		return
	}
	status := transform(machineStatus)
	a.publisher.Publish(msg.Status, status)
}

func transform(machineStatus MachineStatus) Status {
	return Status{
		CalculatedStatus{},
		machineStatus,
	}
}

// UpdateConfig requests component broadcast current configuration
func (a Asset) UpdateConfig() {
	a.publisher.Publish(msg.Config, a.config())
}

//Config returns the archetypical configuration for the energy storage system asset.
func (a Asset) config() MachineConfig {
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

// MachineStatus is a data structure representing an architypical status
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

// Config differentiates between two types of configurations, static and dynamic
type Config struct {
	mux     *sync.Mutex
	machine MachineConfig
}

// MachineConfig holds the asset configuration parameters
type MachineConfig struct {
	Name    string `json:"Name"`
	BusName string `json:"BusName"`
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

	publisher := msg.NewPublisher(pid)

	status := Status{&sync.Mutex{}, 0, MachineStatus{}}
	config := Config{&sync.Mutex{}, machineConfig}

	return Asset{PID, device, publisher, config}, err

}
