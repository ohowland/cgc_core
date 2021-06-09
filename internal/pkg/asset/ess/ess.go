package ess

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// DeviceController is the hardware abstraction layer
type DeviceController interface {
	ReadDeviceStatus() (MachineStatus, error)
	WriteDeviceControl(MachineControl) error
	Stop() error
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
	return a.config.Static.Name
}

// BusName is a getter for the asset's connected Bus
func (a Asset) BusName() string {
	return a.config.Static.BusName
}

// DeviceController returns the hardware abstraction layer struct
func (a Asset) DeviceController() DeviceController {
	return a.device
}

// Subscribe returns a channel on which the specified topic is broadcast
func (a Asset) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	ch, err := a.publisher.Subscribe(pid, topic)
	return ch, err
}

// Unsubscribe pid from all topic broadcasts
func (a *Asset) Unsubscribe(pid uuid.UUID) {
	a.publisher.Unsubscribe(pid)
}

func (a *Asset) Stop() {
	a.publisher.Stop()
	a.device.Stop()
}

// RequestControl connects the asset control to the read only channel parameter.
func (a *Asset) RequestControl(pid uuid.UUID, ch <-chan msg.Msg) error {
	a.mux.Lock()
	defer a.mux.Unlock()
	// TODO: previous owner needs to stop. how to enforce?
	a.controlOwner = pid
	go a.controlHandler(ch)

	return nil
}

// UpdateStatus requests a physical device read, then broadcasts results
func (a Asset) UpdateStatus() {
	machineStatus, err := a.device.ReadDeviceStatus()
	if err != nil {
		// Read Error Handler Path
		return
	}
	calcStatus := a.calculateStatus(machineStatus)
	status := transform(machineStatus, calcStatus)
	a.publisher.Publish(msg.Status, status)
}

// UpdateConfig requests component broadcast current configuration
func (a Asset) UpdateConfig() {
	a.publisher.Publish(msg.Config, a.config)
}

func (a Asset) calculateStatus(ms MachineStatus) CalculatedStatus {
	realPositiveCapacity := calcRealPositiveCapacity(ms, a.config)
	realNegativeCapacity := calcRealNegativeCapacity(ms, a.config)
	storedEnergy := calcStoredEnergy(ms, a.config)
	storedEnergyCapacity := calcStoredEnergyCapacity(ms, a.config)
	energyProductionCost := calcEnergyProductionCost(ms, a.config)
	energyConsumptionValue := calcEnergyConsumptionValue(ms, a.config)
	capacityCost := calcCapacityCost(ms, a.config)

	return CalculatedStatus{
		realPositiveCapacity,
		realNegativeCapacity,
		storedEnergy,
		storedEnergyCapacity,
		energyProductionCost,
		energyConsumptionValue,
		capacityCost,
	}
}

func calcRealPositiveCapacity(ms MachineStatus, cfg Config) float64 {
	if ms.Faulted {
		return 0.0
	}
	// need to curtail based on SOC
	return cfg.Static.RatedKVA
}

func calcRealNegativeCapacity(ms MachineStatus, cfg Config) float64 {
	if ms.Faulted {
		return 0.0
	}
	// need to curtail based on SOC
	return cfg.Static.RatedKVA
}

func calcStoredEnergy(ms MachineStatus, cfg Config) float64 {
	return ms.SOC * cfg.Static.MeasuredKWH
}

func calcStoredEnergyCapacity(ms MachineStatus, cfg Config) float64 {
	return cfg.Static.MeasuredKWH
}

func calcEnergyProductionCost(ms MachineStatus, cfg Config) float64 {
	return cfg.Static.EnergyProductionCost
}

func calcEnergyConsumptionValue(ms MachineStatus, cfg Config) float64 {
	return cfg.Static.EnergyConsumptionValue
}

func calcCapacityCost(ms MachineStatus, cfg Config) float64 {
	return 1
}

func transform(ms MachineStatus, cs CalculatedStatus) Status {
	return Status{
		cs,
		ms,
	}
}

// Shutdown instructs the asset to cleanup all resources
func (a Asset) Shutdown(wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()
	return a.device.Stop()
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

// Status wraps MachineStatus with mutex and state metadata
type Status struct {
	Calc    CalculatedStatus `json:"CalculatedStatus"`
	Machine MachineStatus    `json:"MachineStatus"`
}

// CalculatedStatus is a data structure representing asset state information
// that is calculated from data read into the archetype ess.
type CalculatedStatus struct {
	RealPositiveCapacity   float64 `json:"RealPositiveCapacity"`
	RealNegativeCapacity   float64 `json:"RealNegativeCapacity"`
	StoredEnergy           float64 `json:"StoredEnergy"`
	StoredEnergyCapacity   float64 `json:"StoredEnergyCapacity"`
	EnergyProductionCost   float64 `json:"EnergyProductionCost"`
	EnergyConsumptionValue float64 `json:"EnergyConsumptionValue"`
	CapacityCost           float64 `json:"CapacityCost"`
}

// MachineStatus is a data structure representing an architypical ESS status
type MachineStatus struct {
	KW          float64 `json:"KW"`
	KVAR        float64 `json:"KVAR"`
	Hz          float64 `json:"Hz"`
	Volts       float64 `json:"Volts"`
	SOC         float64 `json:"SOC"`
	Gridforming bool    `json:"Gridforming"`
	Online      bool    `json:"Online"`
	Faulted     bool    `json:"Faulted"`
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
	return s.Calc.RealPositiveCapacity
}

// RealNegativeCapacity returns the asset's operative real negative capacity
func (s Status) RealNegativeCapacity() float64 {
	return s.Calc.RealNegativeCapacity
}

func (s Status) StoredEnergy() float64 {
	return s.Calc.StoredEnergy
}

func (s Status) StoredEnergyCapacity() float64 {
	return s.Calc.StoredEnergyCapacity
}

func (s Status) EnergyProductionCost() float64 {
	return s.Calc.EnergyProductionCost
}

func (s Status) EnergyConsumptionValue() float64 {
	return s.Calc.EnergyConsumptionValue
}

func (s Status) CapacityCost() float64 {
	return s.Calc.CapacityCost
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
	enable bool
}

// Config wraps MachineConfig with mutex a mutex and hides the internal state.
type Config struct {
	Static  StaticConfig  `json:"Static"`
	Dynamic DynamicConfig `json:"Dynamic"`
}

type DynamicConfig struct{}

// StaticConfig holds the ESS asset configuration parameters
type StaticConfig struct {
	Name                   string  `json:"Name"`
	BusName                string  `json:"BusName"`
	RatedKVA               float64 `json:"RatedKVA"`
	RatedKWH               float64 `json:"RatedKWH"`
	MeasuredKWH            float64 `json:"MeasuredKWH"`
	EnergyProductionCost   float64 `json:"EnergyProductionCost"`
	EnergyConsumptionValue float64 `json:"EnergyConsumptionValue"`
	CapacityCost           float64 `json:"CapacityCost"`
}

// New returns a configured Asset
func New(jsonConfig []byte, device DeviceController) (Asset, error) {
	staticConfig := StaticConfig{}
	err := json.Unmarshal(jsonConfig, &staticConfig)
	if err != nil {
		return Asset{}, err
	}

	dynamicConfig := DynamicConfig{}

	pid, err := uuid.NewUUID()
	if err != nil {
		return Asset{}, err
	}

	publisher := msg.NewPublisher(pid)
	controlOwner := uuid.UUID{}
	supervisory := SupervisoryControl{false}
	config := Config{staticConfig, dynamicConfig}

	return Asset{
			&sync.Mutex{},
			pid,
			device,
			publisher,
			controlOwner,
			supervisory,
			config},
		err
}
