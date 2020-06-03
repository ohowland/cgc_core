package manualdispatch

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch/calculatedstatus"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// ManualDispatch is the core datastructure
type ManualDispatch struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	publisher   *msg.PubSub
	calcStatus  *calculatedstatus.CalculatedStatus
	memberState map[uuid.UUID]State
}

type State struct {
	Status  interface{}
	Control interface{}
	Config  interface{}
}

// New returns a configured ManualDispatch struct
func New(configPath string) (*ManualDispatch, error) {
	pid, err := uuid.NewUUID()
	pub := msg.NewPublisher(uuid.UUID{})
	calcStatus, err := calculatedstatus.NewCalculatedStatus()
	memberState := make(map[uuid.UUID]State)
	return &ManualDispatch{
			&sync.Mutex{},
			pid,
			pub,
			&calcStatus,
			memberState,
		},
		err
}

func (d *ManualDispatch) StartProcess(ch <-chan msg.Msg) error {
	go d.Process(ch)
	return nil
}

func (d *ManualDispatch) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return d.publisher.Subscribe(pid, topic)
}

func (d *ManualDispatch) Unsubscribe(pid uuid.UUID) {
	d.publisher.Unsubscribe(pid)
}

func (d *ManualDispatch) Process(ch <-chan msg.Msg) {
	ticker := time.NewTicker(100 * time.Millisecond)
loop:
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				break loop
			}
			state := State{Status: m.Payload()}
			// lock mutex?
			d.memberState[m.PID()] = state
			fmt.Println(d.memberState)
		case <-ticker.C:
			fmt.Println("[Dispatch] Write")
		}
	}
	fmt.Println("[Dispatch] Goroutine Shutdown")
}

// DropAsset ...
func (d *ManualDispatch) DropAsset(pid uuid.UUID) error {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.calcStatus.DropAsset(pid)
	delete(d.memberState, pid)
	return nil
}

// GetControl ...
func (d ManualDispatch) GetControl(pid uuid.UUID) (interface{}, bool) {
	state, ok := d.memberState[pid]
	return state.Control, ok
}

// GetStatus ...
func (d ManualDispatch) GetStatus(pid uuid.UUID) (interface{}, bool) {
	state, ok := d.memberState[pid]
	return state.Status, ok
}

func (d ManualDispatch) PID() uuid.UUID {
	return d.pid
}
