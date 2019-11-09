package virtualacbus

import (
	"io/ioutil"
	"log"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
)

type VirtualACBus struct {
	pid              uuid.UUID
	busObserver      chan Source
	assetObserver    chan Source
	connectedSources map[uuid.UUID]Source
	gridformer       Source
}

type Source struct {
	PID         uuid.UUID
	Hz          float64
	Volt        float64
	KW          float64
	KVAR        float64
	Gridforming bool
}

// Observers contains the virtual system communication interface
type Observers struct {
	AssetObserver chan<- Source
	BusObserver   <-chan Source
}

func (b VirtualACBus) ReadDeviceStatus() (acbus.RelayStatus, error) {
	return acbus.RelayStatus{
		Hz:   b.gridformer.Hz,
		Volt: b.gridformer.Volt,
	}, nil
}

func (b VirtualACBus) PID() uuid.UUID {
	return b.pid
}

func (b VirtualACBus) GetBusObservers() Observers {
	return Observers{
		AssetObserver: b.AssetObserver(),
		BusObserver:   b.BusObserver(),
	}
}

func (b VirtualACBus) Hz() float64 {
	return b.gridformer.Hz
}

func (b VirtualACBus) Volt() float64 {
	return b.gridformer.Volt
}

func (b VirtualACBus) AssetObserver() chan<- Source {
	return b.assetObserver
}

func (b VirtualACBus) BusObserver() <-chan Source {
	return b.busObserver
}

func (b VirtualACBus) gridformerCalcs() Source {
	kwSum := 0.0
	kvarSum := 0.0
	var gridformer Source
	for _, s := range b.connectedSources {
		if s.Gridforming != true {
			kwSum += s.KW
			kvarSum += s.KVAR
		} else {
			gridformer = s
		}
	}
	gridformer.KW = kwSum * -1
	gridformer.KVAR = kvarSum
	return gridformer
}

func (b *VirtualACBus) StartProcess() {
	assetObs := make(chan Source)
	busObs := make(chan Source)
	b.assetObserver = assetObs
	b.busObserver = busObs
	go b.Process()
}

func (b *VirtualACBus) StopProcess() {
	close(b.busObserver)
}

func (b *VirtualACBus) Process() {
	log.Println("[VirtualACBus: Running]")
loop:
	for {
		select {
		case source, ok := <-b.assetObserver:
			if !ok {
				break loop
			}
			b.connectedSources[source.PID] = source

			if source.Gridforming == true {
				b.gridformer = source
			}
		case b.busObserver <- b.gridformerCalcs():
		default:
			var gridformerFound bool
			for _, s := range b.connectedSources {
				if s.Gridforming {
					gridformerFound = true
				}
			}
			if !gridformerFound {
				b.gridformer = Source{}
			}
		}
	}
}

// New returns an initalized VirtualESS Asset; this is part of the Asset interface.
func New(configPath string) (bus.Bus, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return acbus.ACBus{}, err
	}

	id, _ := uuid.NewUUID()
	virtualsystem := VirtualACBus{
		pid:              id,
		busObserver:      nil,
		assetObserver:    nil,
		connectedSources: make(map[uuid.UUID]Source),
		gridformer: Source{
			PID:         id,
			Hz:          0.0,
			Volt:        0.0,
			KW:          0.0,
			KVAR:        0.0,
			Gridforming: true,
		},
	}

	go virtualsystem.StartProcess()
	return acbus.New(jsonConfig, &virtualsystem)
}
