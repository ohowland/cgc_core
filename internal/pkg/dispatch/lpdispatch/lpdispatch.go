package lpdispatch

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	opt "github.com/ohowland/cgc_optimize"
)

type LPDispatch struct {
	mux       *sync.Mutex
	pid       uuid.UUID
	publisher *msg.PubSub
	units     map[uuid.UUID]opt.Unit
	//lp        opt.MipLinearProgram
}

func New(configPath string) (*LPDispatch, error) {
	pid, err := uuid.NewUUID()
	pub := msg.NewPublisher(pid)
	return &LPDispatch{
			&sync.Mutex{},
			pid,
			pub,
			make(map[uuid.UUID]opt.Unit),
		},
		err
}

func (d LPDispatch) PID() uuid.UUID {
	return d.pid
}

func (d *LPDispatch) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return d.publisher.Subscribe(pid, topic)
}

func (d *LPDispatch) Unsubscribe(pid uuid.UUID) {
	d.publisher.Unsubscribe(pid)
}

func (d *LPDispatch) StartProcess(ch <-chan msg.Msg) error {
	log.Println("[LP Dispatch] Starting")
	go d.Process(ch)
	return nil
}

func (d *LPDispatch) Process(ch <-chan msg.Msg) {
	ticker := time.NewTicker(5000 * time.Millisecond)
loop:
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				log.Println("[LP Dispatch] disconnected from bus graph")
				break loop
			}
			d.ingress(m)
		case <-ticker.C:
			//d.runSolver()

		}
	}
}

// ingress updates the state of the linear program
func (d *LPDispatch) ingress(m msg.Msg) {
	switch m.Topic() {
	case msg.Status:
		unit := asUnit(m.PID(), m.Payload())
		d.units[m.PID()] = unit
	case msg.Config:
	default:
	}
}
