package virtualdcbus

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/bus/dc"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// VirtualDCBus is the top level data structure for virtual bus.
type VirtualDCBus struct {
	mux           *sync.Mutex
	pid           uuid.UUID
	assetReciever chan asset.VirtualDCStatus
	inbox         chan msg.Msg
	members       map[uuid.UUID]bool
	stopProcess   chan bool
}

// New returns an initalized VirtualDCBus Asset; this is part of the Asset interface.
func New(configPath string) (dc.Bus, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return dc.Bus{}, err
	}

	id, _ := uuid.NewUUID()
	virtualsystem := VirtualDCBus{
		mux:           &sync.Mutex{},
		pid:           id,
		assetReciever: make(chan asset.VirtualDCStatus),
		inbox:         make(chan msg.Msg),
		members:       make(map[uuid.UUID]bool),
		stopProcess:   make(chan bool),
	}

	return dc.New(jsonConfig, &virtualsystem)
}

// PID is an accessor for the process id
func (b VirtualDCBus) PID() uuid.UUID {
	return b.pid
}

// Volts is an accessor for the bus voltage
func (b VirtualDCBus) Volts() float64 {
	status := <-b.assetReciever
	return status.Volts()
}

// AddMember joins a virtual asset to the virtual bus.
func (b *VirtualDCBus) AddMember(a asset.VirtualDCAsset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender := a.LinkToBus(b.assetReciever)
	b.members[a.PID()] = true

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan asset.VirtualDCStatus, inbox chan<- msg.Msg) {
		for status := range assetSender {
			inbox <- msg.New(pid, msg.Status, status)
		}
		b.removeMember(a.PID())                       // on channel close revoke membership
		inbox <- msg.New(pid, msg.Status, Template{}) // and clear contribuiton.
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

// removeMember removes a virtual asset from the virtual bus
func (b *VirtualDCBus) removeMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.members, pid)

	if len(b.members) < 1 { // if this is the first member, start the bus process.
		b.stopProcess <- true
	}
}

// StopProcess terminates the virtual bus process loop.
// This is used for controlled shutdowns.
func (b *VirtualDCBus) StopProcess() {
	b.mux.Lock()
	defer b.mux.Unlock()
	allPIDs := make([]uuid.UUID, len(b.members))

	for pid := range b.members {
		allPIDs = append(allPIDs, pid)
	}

	for _, pid := range allPIDs {
		delete(b.members, pid)
	}

	b.stopProcess <- true
}

type assetMap map[uuid.UUID]asset.VirtualDCStatus

// Process is the main process loop for the virtual bus.
// This loop is responsible for aggregating virtual component data,
// and calculating the power balance (swing load) for gridformer.
func (b *VirtualDCBus) Process() {
	defer close(b.assetReciever)
	log.Println("[VirtualBus] Starting")
	memberStatus := make(map[uuid.UUID]asset.VirtualDCStatus)
loop:
	for {
		select {
		case msg, ok := <-b.inbox:
			if !ok {
				break loop
			} else {
				memberStatus = b.processMsg(msg, memberStatus)
			}
		case b.assetReciever <- busPowerBalance(memberStatus):

		case <-b.stopProcess:
			break loop
		}
	}
	log.Println("[VirtualBus] Goroutine Shutdown")
}

func (b *VirtualDCBus) processMsg(msg msg.Msg, memberStatus assetMap) assetMap {
	if b.hasMember(msg.PID()) {
		memberStatus = aggregateStatus(msg, memberStatus)
	} else if _, ok := memberStatus[msg.PID()]; ok { // if non-member, remove stale data -
		delete(memberStatus, msg.PID()) // this is currently the mechanism to remove old data.
	}
	return memberStatus
}

func (b *VirtualDCBus) hasMember(pid uuid.UUID) bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	_, ok := b.members[pid]
	return ok
}

// updateAggregate manages the aggregation of asset status.
func aggregateStatus(msg msg.Msg, agg assetMap) assetMap {
	agg[msg.PID()] = msg.Payload().(asset.VirtualDCStatus)
	return agg
}

// Template data structure is used to hold the essential data for the
// grid forming device.
type Template struct {
	kW         float64
	volts      float64
	gridformer bool
}

// KW is an accessor for real power
func (v Template) KW() float64 {
	return v.kW
}

func (v Template) KVAR() float64 {
	return 0
}

// Volt is an accessor for voltage
func (v Template) Volts() float64 {
	return v.volts
}

// Gridforming is required to meet the asset.VirtualDCStatus interface.
func (v Template) Gridforming() bool {
	return v.gridformer
}

func newTemplate(a asset.VirtualDCStatus) Template {
	return Template{
		kW:         a.KW(),
		volts:      a.Volts(),
		gridformer: a.Gridforming(),
	}
}

func busPowerBalance(agg map[uuid.UUID]asset.VirtualDCStatus) Template {
	kwSum := 0.0
	gridformerStatus := Template{}
	for _, assetStatus := range agg {
		if assetStatus.Gridforming() {
			gridformerStatus = newTemplate(assetStatus)
		} else {
			kwSum += assetStatus.KW()
		}

	}
	gridformerStatus.kW = kwSum * -1
	return gridformerStatus
}
