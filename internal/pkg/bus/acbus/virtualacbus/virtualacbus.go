package virtualacbus

import (
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
)

type VirtualACBus struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	busObserver chan asset.VirtualAssetStatus
	members     map[uuid.UUID]<-chan asset.VirtualAssetStatus
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string) (bus.Bus, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return acbus.ACBus{}, err
	}

	id, _ := uuid.NewUUID()
	virtualsystem := VirtualACBus{
		mux:         &sync.Mutex{},
		pid:         id,
		busObserver: make(chan asset.VirtualAssetStatus),
		members:     make(map[uuid.UUID]<-chan asset.VirtualAssetStatus),
	}

	return acbus.New(jsonConfig, &virtualsystem)
}

func (b VirtualACBus) ReadRelayStatus() (acbus.RelayStatus, error) {
	status := <-b.busObserver
	return acbus.NewRelayStatus(
		status.Hz(),
		status.Volt(),
	), nil
}

func (b VirtualACBus) PID() uuid.UUID {
	return b.pid
}

func (b *VirtualACBus) AddMember(a asset.VirtualAsset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.members[a.PID()] = a.LinkToBus(b.busObserver)
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
	for pid := range b.members {
		delete(b.members, pid)
	}
}

func (b *VirtualACBus) Process() {
	defer close(b.busObserver)
	agg := make(map[uuid.UUID]asset.VirtualAssetStatus)
loop:
	for {
		b.mux.Lock()
		members := b.members
		b.mux.Unlock()
		if len(members) == 0 { // if there are no members, end the bus process.
			break loop
		}
		for pid, member := range members {
			select {
			case assetStatus, ok := <-member:
				if !ok {
					log.Println("it's bad") //testing!
					b.RemoveMember(pid)
					delete(agg, pid)
				} else {
					agg[pid] = assetStatus
				}
			case b.busObserver <- busPowerBalance(agg):
			}
		}
	}
}

type virtualGridFormer struct {
	kW   float64
	kVAR float64
	hz   float64
	volt float64
}

func (v virtualGridFormer) KW() float64 {
	return v.kW
}
func (v virtualGridFormer) KVAR() float64 {
	return v.kVAR

}
func (v virtualGridFormer) Hz() float64 {
	return v.hz
}

func (v virtualGridFormer) Volt() float64 {
	return v.volt
}

func (v virtualGridFormer) Gridforming() bool {
	return true
}

func newVirtualGridFormer(a asset.VirtualAssetStatus) virtualGridFormer {
	return virtualGridFormer{
		kW:   a.KW(),
		kVAR: a.KVAR(),
		hz:   a.Hz(),
		volt: a.Volt(),
	}
}

func busPowerBalance(agg map[uuid.UUID]asset.VirtualAssetStatus) virtualGridFormer {
	kwSum := 0.0
	kvarSum := 0.0
	var gridformerStatus virtualGridFormer
	for _, assetStatus := range agg {
		if assetStatus.Gridforming() == false {
			kwSum += assetStatus.KW()
			kvarSum += assetStatus.KVAR()
		} else {
			gridformerStatus = newVirtualGridFormer(assetStatus)
		}

	}
	gridformerStatus.kW = kwSum * -1
	gridformerStatus.kVAR = kvarSum

	return gridformerStatus
}
