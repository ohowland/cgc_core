package asset

import "time"

// ActorProcess is the wrapper around all Asset objects that communicate with physical devices
type ActorProcess interface {
	Initialize() error
	Run() error
	Stop() error
}

// Scheduler sends recurring messages to AssetActors
type Scheduler struct {
	ch      chan string
	targets []Target
}

// Run the scheduler
func (s *Scheduler) Run() {
	s.ch = make(chan string)
	go runScheduler(s)
	return
}

// Stop the scheduler
func (s *Scheduler) Stop() {
	s.ch <- "stop"
	close(s.ch)
	return
}

// AddTarget appends a Target to the schedulers messaging list
func (s *Scheduler) AddTarget(t Target) {
	s.targets = append(s.targets, t)
}

// NewScheduler returns an initalized Scheduler
func NewScheduler() Scheduler {
	return Scheduler{}
}

func runScheduler(s *Scheduler) {
	for {
		select {
		case _ = <-s.ch:
			return
		default:
			for _, target := range s.targets {
				target.send()
			}
		}
	}
}

// Target is the reciever of a message at a specific rate
type Target struct {
	target  chan interface{}
	message interface{}
	rate    time.Duration
	last    time.Time
}

func (t *Target) send() {
	if time.Since(t.last) > t.rate {
		t.target <- t.message
		t.last = time.Now()
	}
	return
}

// NewTarget returns an initialized Target
func NewTarget(target chan interface{}, message interface{}, rateMillis int) Target {
	return Target{target, message, time.Duration(rateMillis) * time.Millisecond, time.Now()}
}

// ActorProcessState encodes available actor states
type ActorProcessState int

const (
	uninitialized ActorProcessState = iota
	initialized   ActorProcessState = iota
	running       ActorProcessState = iota
	stopped       ActorProcessState = iota
)
