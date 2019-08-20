package asset

import (
	"time"
)

// Process is a wrapper around the device object.
type Process struct {
	asset     *Asset
	scheduler Scheduler
	inbox     chan interface{}
	state     ProcessState
}

// ProcessState encodes available actor states
type ProcessState int

const (
	uninitialized ProcessState = iota
	initialized   ProcessState = iota
	running       ProcessState = iota
	stopped       ProcessState = iota
)

// UpdateStatus requests process to perform a device status read
type UpdateStatus struct{}

// WriteControl requests process to perform a device control write
type WriteControl struct{}

// Start the process
type Start struct{}

// Stop the actor
type Stop struct{}

// StartProcess spins up a Process
func StartProcess(a Asset) chan interface{} {
	proc := Process{asset: &a, inbox: nil, state: uninitialized}
	inbox := proc.initialize()
	go proc.run()
	return inbox
}

// Initialize fulfills Initialize() in the Process Interface. The Process is initilized and put in the Run state.
func (p *Process) initialize() chan interface{} {

	p.inbox = make(chan interface{})
	p.state = initialized

	p.scheduler = NewScheduler()
	read := NewTarget(p.inbox, UpdateStatus{}, 1000)
	write := NewTarget(p.inbox, WriteControl{}, 1000)
	p.scheduler.addTarget(read)
	p.scheduler.addTarget(write)

	return p.inbox
}

// Run fulfills the Process Interface. In this state the Process starts the scheduler and recieves messages.
func (p *Process) run() error {
	p.scheduler.run()
	p.state = running
	for msg := range p.inbox {
		switch msg.(type) {

		case UpdateStatus:
			// mutex?
			asset := *p.asset
			asset.UpdateStatus()

		case WriteControl:
			// mutex?
			asset := *p.asset
			asset.WriteControl()

		case Stop:
			go p.stop()
			return nil
		}
	}
	return nil
}

// Stop fulfills the Process Interface. In this state the Process pauses the scheduler.
func (p *Process) stop() error {
	p.scheduler.stop()
	p.state = stopped
	for msg := range p.inbox {
		switch msg.(type) {
		case Start:
			go p.run()
			return nil
		}
	}
	return nil
}

// Scheduler sends recurring messages to Process
type Scheduler struct {
	ch      chan string
	targets []Target
}

// NewScheduler returns an initalized Scheduler
func NewScheduler() Scheduler {
	return Scheduler{}
}

// Run the scheduler
func (s *Scheduler) run() {
	s.ch = make(chan string)
	go runScheduler(s)
}

// Stop the scheduler
func (s *Scheduler) stop() {
	close(s.ch)
}

// AddTarget appends a Target to the schedulers messaging list
func (s *Scheduler) addTarget(t Target) {
	s.targets = append(s.targets, t)
}

func runScheduler(s *Scheduler) {
	for {
		select {
		case _, ok := <-s.ch:
			if !ok {
				return
			}

		default:
			for i := range s.targets {
				s.targets[i].send()
			}
		}
	}
}

// Target is the reciever of a message at a specific rate
type Target struct {
	target chan interface{}
	msg    interface{}
	rate   time.Duration
	last   time.Time
}

// NewTarget returns an initialized Target
func NewTarget(target chan interface{}, message interface{}, rateMillis int) Target {
	return Target{target, message, time.Duration(rateMillis) * time.Millisecond, time.Now()}
}

// Sends msg to target channel
func (t *Target) send() {
	if time.Since(t.last) > t.rate {
		select {
		case t.target <- t.msg:
			t.last = time.Now()
		default:
		}
	}
}
