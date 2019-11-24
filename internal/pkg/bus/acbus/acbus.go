/*
acbus.go Representation of a single AC bus. Data structures that implement the Asset interface
may join as members. Members report asset.AssetStatus, which is aggregated by the bus.
*/

package acbus

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type ACBus struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	relay       Relayer
	inbox       chan Msg
	members     map[uuid.UUID]bool
	config      Config
	stopProcess chan bool
}

type Status struct {
	capacity   Capacity
	renewables Renewables
}

type Capacity struct {
	RealPositiveCapacity float64
	RealNegativeCapacity float64
}

type Renewables struct {
	RE_KW float64
}

type Config struct {
	Name      string  `json:"Name"`
	RatedVolt float64 `json:"RatedVolt"`
	RatedHz   float64 `json:"RatedHz"`
}

func New(jsonConfig []byte, relay Relayer) (ACBus, error) {

	config := Config{}
	err := json.Unmarshal(jsonConfig, &config)
	if err != nil {
		return ACBus{}, err
	}

	PID, err := uuid.NewUUID()
	if err != nil {
		return ACBus{}, err
	}

	inbox := make(chan Msg)
	//dispatch := make(chan Status)
	stopProcess := make(chan bool)
	members := make(map[uuid.UUID]bool)

	return ACBus{
		&sync.Mutex{},
		PID,
		relay,
		inbox,
		members,
		config,
		stopProcess}, nil
}

func (b ACBus) Name() string {
	return b.config.Name
}

func (b ACBus) PID() uuid.UUID {
	return b.pid
}

func (b ACBus) Relayer() Relayer {
	return b.relay
}

func (b *ACBus) AddMember(a asset.Asset) {
	b.mux.Lock()
	defer b.mux.Unlock()
	assetSender := a.Subscribe(b.pid)
	b.members[a.PID()] = true

	// aggregate messages from assets into the busReciever channel, which is read in the Process loop.
	go func(pid uuid.UUID, assetSender <-chan interface{}, inbox chan<- Msg) {
		for status := range assetSender {
			inbox <- Msg{pid, status}
		}
	}(a.PID(), assetSender, b.inbox)

	if len(b.members) == 1 { // if this is the first member, start the bus process.
		go b.Process()
	}
}

func (b *ACBus) RemoveMember(pid uuid.UUID) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.members, pid)
}

func (b *ACBus) StopProcess() {
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

// Process aggregates information from members, while members exist.
func (b *ACBus) Process() {
	defer close(b.inbox)
	log.Println("ACBus Process: Loop Started")
	memberStatus := make(map[uuid.UUID]Status)
loop:
	for {
		select {
		case msg, ok := <-b.inbox:
			if !ok {
				b.RemoveMember(msg.PID())
				delete(memberStatus, msg.PID())
			} else {
				memberStatus = b.updateMemberStatus(msg, memberStatus)
			}
		//case b.dispatch <- b.updateBusStatus(memberStatus):
		case <-b.stopProcess:
			break loop
		}
	}
	log.Println("ACBus Process: Loop Stopped")
}

func (b ACBus) hasMember(pid uuid.UUID) bool {
	return b.members[pid] == true
}

func (b ACBus) updateMemberStatus(msg Msg, memberStatus map[uuid.UUID]Status) map[uuid.UUID]Status {
	if b.hasMember(msg.PID()) {
		status := memberStatus[msg.PID()]
		switch p := msg.payload.(type) {
		case asset.Capacity:
			status.capacity = Capacity{
				RealPositiveCapacity: p.RealNegativeCapacity(),
				RealNegativeCapacity: p.RealNegativeCapacity(),
			}
		case asset.Renewable:
			status.renewables = Renewables{
				RE_KW: p.RE_KW(),
			}
		}
		memberStatus[msg.PID()] = status
	}
	return memberStatus
}

func (b ACBus) updateBusStatus(memberStatus map[uuid.UUID]Status) Status {
	return Status{
		capacity: aggregateCapacity(memberStatus),
	}
}

func aggregateCapacity(memberStatus map[uuid.UUID]Status) Capacity {
	var realPositiveCapacity float64
	var realNegativeCapacity float64
	for _, status := range memberStatus {
		realPositiveCapacity += status.capacity.RealPositiveCapacity
		realNegativeCapacity += status.capacity.RealNegativeCapacity
	}
	return Capacity{
		realPositiveCapacity,
		realNegativeCapacity,
	}
}

// UpdateRelayer requests a physical device read, then updates MachineStatus field.
func (b ACBus) UpdateRelayer() (RelayStatus, error) {
	relayStatus, err := b.relay.ReadDeviceStatus()
	if err != nil {
		return EmptyRelayStatus{}, err
	}
	return relayStatus, nil
}

type Msg struct {
	sender  uuid.UUID
	payload interface{}
}

func (v Msg) PID() uuid.UUID {
	return v.sender
}

func (v Msg) Payload() interface{} {
	return v.payload
}

type Relayer interface {
	ReadDeviceStatus() (RelayStatus, error)
}

type RelayStatus interface {
	Hz() float64
	Volt() float64
}

type EmptyRelayStatus struct{}

func (s EmptyRelayStatus) Hz() float64   { return 0 }
func (s EmptyRelayStatus) Volt() float64 { return 0 }
