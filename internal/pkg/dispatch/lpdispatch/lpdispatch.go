package lpdispatch

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	opt "github.com/ohowland/cgc_optimize"
	"github.com/ohowland/highs"
)

type LPDispatch struct {
	mux       *sync.Mutex
	pid       uuid.UUID
	publisher *msg.PubSub
	units     map[uuid.UUID]opt.Unit
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
			d.runSolver()
		}
	}
}

// ingress updates the state of the linear program
func (d *LPDispatch) ingress(m msg.Msg) {
	switch m.Topic() {
	case msg.Status:
		unit := BuildUnit(m.PID(), m.Payload())
		d.units[m.PID()] = unit
	default:
	}
}

func (d *LPDispatch) runSolver() ([]float64, error) {
	units := make([]opt.Unit, 0)
	for _, u := range d.units {
		fmt.Printf("%+v\n", u)
		units = append(units, u)
	}
	grp := opt.NewGroup(units...)

	netload := 5.0
	grp.NewConstraint(opt.NetLoadConstraint(&grp, netload))

	s, err := highs.New(
		grp.CostCoefficients(),
		grp.Bounds(),
		grp.Constraints(),
		[]int{})

	if err != nil {
		return []float64{}, err
	}

	s.SetObjectiveSense(highs.Minimize)
	s.RunSolver()
	sol := s.PrimalColumnSolution()
	log.Println(sol)

	return sol, nil
}
