package virtualacbus

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
)

type VirtualACBus struct {
	mux           *sync.Mutex
	pid           uuid.UUID
	assetReciever chan asset.VirtualStatus
	inbox         chan Msg
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
		inbox:         make(chan Msg),
		members:       make(map[uuid.UUID]bool),
		stopProcess:   make(chan bool),
	}

	return acbus.New(jsonConfig, &virtualsystem)
}

func (b VirtualACBus) PID() uuid.UUID {
	return b.pid
}

func (b VirtualACBus) ReadRelayStatus() (acbus.RelayStatus, error) {
	status := <-b.assetReciever
	return acbus.NewRelayStatus(
		status.Hz(),
		status.Volt(),
	), nil
}

func (b *VirtualACBus) AddMember(a asset.VirtualAsset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender := a.LinkToBus(b.assetReciever)
	b.members[a.PID()] = true

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan asset.VirtualStatus, inbox chan<- Msg) {
		for status := range assetSender {
			inbox <- Msg{pid, status}
		}
		inbox <- Msg{pid, Template{}} // on close clear contribution with default status.
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

func (b *VirtualACBus) RemoveMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.members, pid)
}

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

func (b *VirtualACBus) Process() {
	defer close(b.assetReciever)
	log.Println("VirtualBus Process: Loop Started")
	agg := make(map[uuid.UUID]asset.VirtualStatus)
loop:
	for {
		select {
		case msg, ok := <-b.inbox:
			if !ok {
				b.RemoveMember(msg.PID())
				delete(agg, msg.PID())
			} else {
				agg = b.updateAggregates(msg, agg)
			}
		case b.assetReciever <- busPowerBalance(agg):
		case <-b.stopProcess:
			break loop
		}
	}
	log.Println("VirtualBus Process: Loop Stopped")
}

func (b *VirtualACBus) updateAggregates(msg Msg,
	agg map[uuid.UUID]asset.VirtualStatus) map[uuid.UUID]asset.VirtualStatus {
	b.mux.Lock()
	defer b.mux.Unlock()
	if _, ok := b.members[msg.PID()]; ok { // filter for member pid
		agg[msg.PID()] = msg.Status()
	}

	return agg
}

type Msg struct {
	sender uuid.UUID
	status asset.VirtualStatus
}

func (v Msg) PID() uuid.UUID {
	return v.sender
}

func (v Msg) Status() asset.VirtualStatus {
	return v.status
}

type Template struct {
	kW   float64
	kVAR float64
	hz   float64
	volt float64
}

func (v Template) KW() float64 {
	return v.kW
}
func (v Template) KVAR() float64 {
	return v.kVAR

}
func (v Template) Hz() float64 {
	return v.hz
}

func (v Template) Volt() float64 {
	return v.volt
}

func (v Template) Gridforming() bool {
	return true
}

func newTemplate(a asset.VirtualStatus) Template {
	return Template{
		kW:   a.KW(),
		kVAR: a.KVAR(),
		hz:   a.Hz(),
		volt: a.Volt(),
	}
}

func busPowerBalance(agg map[uuid.UUID]asset.VirtualStatus) Template {
	kwSum := 0.0
	kvarSum := 0.0
	var gridformerStatus Template
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
