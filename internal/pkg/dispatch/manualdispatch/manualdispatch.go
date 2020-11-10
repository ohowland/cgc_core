package manualdispatch

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc_core/internal/pkg/asset/grid"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch/model"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

// ManualDispatch is the core datastructure
type ManualDispatch struct {
	mux         *sync.Mutex
	pid         uuid.UUID
	publisher   *msg.PubSub
	model       *model.Model
	memberState map[uuid.UUID]dispatch.State
}

// New returns a configured ManualDispatch struct
func New(configPath string) (*ManualDispatch, error) {
	pid, err := uuid.NewUUID()
	pub := msg.NewPublisher(uuid.UUID{})
	model, err := model.NewModel()
	memberState := make(map[uuid.UUID]dispatch.State)
	return &ManualDispatch{
			&sync.Mutex{},
			pid,
			pub,
			&model,
			memberState,
		},
		err
}

func (d *ManualDispatch) StartProcess(ch <-chan msg.Msg) error {
	log.Println("[ManualDispatch] Starting")
	go d.Process(ch)
	return nil
}

func (d *ManualDispatch) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return d.publisher.Subscribe(pid, topic)
}

func (d *ManualDispatch) Unsubscribe(pid uuid.UUID) {
	d.publisher.Unsubscribe(pid)
}

// Process is the main loop
func (d *ManualDispatch) Process(ch <-chan msg.Msg) {
	ticker := time.NewTicker(5000 * time.Millisecond)
loop:
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				// link to bus graph lost
				log.Println("[Dispatch] disconnected from bus graph")
				break loop
			}
			d.ingress(m)
		case <-ticker.C:
			d.model.Update(d.memberState)
			// d.optimization.Run()
			// d.stateMachine.Run()
			if m, ok := d.gridRunMsg(); ok {
				//log.Println("[Dispatch] Write", m)
				d.publisher.Publish(msg.Control, m)
			} else {
				log.Println("[Dispatch] No Grid Asset Found")
			}

			if m, ok := d.feederRunMsg(); ok {
				d.publisher.Publish(msg.Control, m)
			} else {
				log.Println("[Dispatch] No Feeder Asset Found")
			}
		}
	}
	log.Println("[Dispatch] Goroutine Shutdown")
}

func (d *ManualDispatch) ingress(m msg.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()

	switch m.Topic() {
	case msg.Status:
		state := d.MemberState(m.PID())
		state.Status = m.Payload()
		d.memberState[m.PID()] = state

	case msg.Config:
		state := d.MemberState(m.PID())
		state.Control = m.Payload()
		d.memberState[m.PID()] = state

	case msg.Control:
		state := d.MemberState(m.PID())
		state.Config = m.Payload()
		d.memberState[m.PID()] = state
	}
}

// MemberState returns state (status, config, control) associated with the PID
func (d ManualDispatch) MemberState(pid uuid.UUID) dispatch.State {
	if state, ok := d.memberState[pid]; ok {
		return state
	}
	return dispatch.State{}
}

/* THIS IS TEMPORARY */
func (d *ManualDispatch) gridRunMsg() (msg.Msg, bool) {
	for pid, state := range d.memberState {
		_, ok := state.Status.(grid.Status)
		if ok {
			control := grid.MachineControl{CloseIntertie: true}
			m := msg.New(pid, msg.Control, control)
			return m, true
		}
	}
	return msg.Msg{}, false
}

func (d *ManualDispatch) feederRunMsg() (msg.Msg, bool) {
	for pid, state := range d.memberState {
		_, ok := state.Status.(feeder.Status)
		if ok {
			control := feeder.MachineControl{CloseFeeder: true}
			m := msg.New(pid, msg.Control, control)
			return m, true
		}
	}
	return msg.Msg{}, false
}

// DropAsset ...
func (d *ManualDispatch) DropAsset(pid uuid.UUID) error {
	d.mux.Lock()
	defer d.mux.Unlock()
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

// PID ...
func (d ManualDispatch) PID() uuid.UUID {
	return d.pid
}
