package virtual

import (
	"log"
	"time"

	"github.com/google/uuid"
)

const (
	queueSize = 50
)

type SystemModel struct {
	ReportLoad chan SourceLoad
	SwingLoad  chan Load
	loads      map[uuid.UUID]Load
	stop       chan bool
}

type SourceLoad struct {
	ID   uuid.UUID
	Load Load
}

type Load struct {
	KW   float64
	KVAR float64
}

func NewVirtualSystemModel() *SystemModel {
	return &SystemModel{
		ReportLoad: make(chan SourceLoad, queueSize),
		SwingLoad:  make(chan Load),
		loads:      make(map[uuid.UUID]Load),
		stop:       make(chan bool, 1),
	}
}

func (s SystemModel) calcSwingLoad() Load {
	kwSum := 0.0
	kvarSum := 0.0
	for _, l := range s.loads {
		kwSum += l.KW
		kvarSum += l.KVAR
	}
	return Load{
		KW:   kwSum,
		KVAR: kvarSum,
	}
}

func (s *SystemModel) RunVirtualSystem() {
	log.Println("[VirtualSystemModel: Running]")
	for {
		select {
		case v := <-s.ReportLoad:
			s.loads[v.ID] = v.Load
			//log.Printf("[VirtualSystemModel: Reported Load %v]\n", v)
		case s.SwingLoad <- s.calcSwingLoad():
			log.Printf("[VirtualSystemModel: Swing Load %v]\n", s.calcSwingLoad())
		case <-s.stop:
			//log.Println("[VirtualSystemModel: stopping]")
			return
		default:
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	}
}

func (s *SystemModel) StopVirtualSystem() {
	s.stop <- true
}
