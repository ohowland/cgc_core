package virtualacbus

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
)

// VirtualACBus is the top level data structure for virtual bus.
type VirtualACBus struct {
	mux           *sync.Mutex
	pid           uuid.UUID
	assetReciever chan asset.VirtualStatus
	inbox         chan asset.Msg
	members       map[uuid.UUID]bool
	stopProcess   chan bool
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string) (acbus.ACBus, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return acbus.ACBus{}, err
	}

	id, _ := uuid.NewUUID()
	virtualsystem := VirtualACBus{
		mux:           &sync.Mutex{},
		pid:           id,
		assetReciever: make(chan asset.VirtualStatus),
		inbox:         make(chan asset.Msg),
		members:       make(map[uuid.UUID]bool),
		stopProcess:   make(chan bool),
	}

	return acbus.New(jsonConfig, &virtualsystem, nil)
}

// PID is an accessor for the process id
func (b VirtualACBus) PID() uuid.UUID {
	return b.pid
}

// Hz is an accessor for the bus frequency
func (b VirtualACBus) Hz() float64 {
	status := <-b.assetReciever
	return status.Hz()
}

// Volt is an accessor for the bus voltage
func (b VirtualACBus) Volt() float64 {
	status := <-b.assetReciever
	return status.Volt()
}

// AddMember joins a virtual asset to the virtual bus.
func (b *VirtualACBus) AddMember(a asset.VirtualAsset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender := a.LinkToBus(b.assetReciever)
	b.members[a.PID()] = true

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan asset.VirtualStatus, inbox chan<- asset.Msg) {
		for status := range assetSender {
			inbox <- asset.NewMsg(pid, status)
		}
		b.removeMember(a.PID())                // on channel close revoke membership
		inbox <- asset.NewMsg(pid, Template{}) // and clear contribuiton.
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

// removeMember removes a virtual asset from the virtual bus
func (b *VirtualACBus) removeMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.members, pid)
}

// StopProcess terminates the virtual bus process loop.
// This is used for controlled shutdowns.
func (b *VirtualACBus) StopProcess() {
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

type assetMap map[uuid.UUID]asset.VirtualStatus

// Process is the main process loop for the virtual bus.
// This loop is responsible for aggregating virtual component data,
// and calculating the power balance (swing load) for gridformer.
func (b *VirtualACBus) Process() {
	defer close(b.assetReciever)
	log.Println("VirtualBus Process: Loop Started")
	memberStatus := make(map[uuid.UUID]asset.VirtualStatus)
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
	log.Println("VirtualBus Process: Loop Stopped")
}

func (b *VirtualACBus) processMsg(msg asset.Msg, memberStatus assetMap) assetMap {
	if b.hasMember(msg.PID()) {
		memberStatus = aggregateStatus(msg, memberStatus)
	} else if _, ok := memberStatus[msg.PID()]; ok { // if non-member, remove stale data -
		delete(memberStatus, msg.PID()) // this is currently the mechanism to remove old data.
	}
	return memberStatus
}

func (b *VirtualACBus) hasMember(pid uuid.UUID) bool {
	b.mux.Lock()
	defer b.mux.Unlock()
	_, ok := b.members[pid]
	return ok
}

// updateAggregate manages the aggregation of asset status.
func aggregateStatus(msg asset.Msg, agg assetMap) assetMap {
	agg[msg.PID()] = msg.Payload().(asset.VirtualStatus)
	return agg
}

// Template data structure is used to hold the essential data for the
// grid forming device.
type Template struct {
	kW         float64
	kVAR       float64
	hz         float64
	volt       float64
	gridformer bool
}

// KW is an accessor for real power
func (v Template) KW() float64 {
	return v.kW
}

// KVAR is an accessor for reactive power
func (v Template) KVAR() float64 {
	return v.kVAR

}

// Hz is an accessor for frequency
func (v Template) Hz() float64 {
	return v.hz
}

// Volt is an accessor for voltage
func (v Template) Volt() float64 {
	return v.volt
}

// Gridforming is required to meet the asset.VirtualStatus interface.
func (v Template) Gridforming() bool {
	return v.gridformer
}

func newTemplate(a asset.VirtualStatus) Template {
	return Template{
		kW:         a.KW(),
		kVAR:       a.KVAR(),
		hz:         a.Hz(),
		volt:       a.Volt(),
		gridformer: a.Gridforming(),
	}
}

func busPowerBalance(agg map[uuid.UUID]asset.VirtualStatus) Template {
	kwSum := 0.0
	kvarSum := 0.0
	gridformerStatus := Template{}
	for _, assetStatus := range agg {
		if assetStatus.Gridforming() == false {
			kwSum += assetStatus.KW()
			kvarSum += assetStatus.KVAR()
		} else {
			gridformerStatus = newTemplate(assetStatus)
		}

	}
	gridformerStatus.kW = kwSum * -1
	gridformerStatus.kVAR = kvarSum
	return gridformerStatus
}
