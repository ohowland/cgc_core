package asset

// GridAsset is a datastructure for an Energy Storage System Asset
type GridAsset interface {
	Status() (GridStatus, error)
	Control(GridControl) error
	Config() (GridStaticConfig, error)
}

// GridStatus is a data structure representing an architypical Grid Intertie status
type GridStatus struct {
	Kw   float64
	Kvar float64
}

// GridControl is a data structure representing an architypical Grid Intertie control
type GridControl struct {
	CloseIntertie bool
}

// GridStaticConfig is a data structure representing an architypical Grid Intertie configuration
type GridStaticConfig struct {
	Name      string
	KwRated   float64
	KvarRated float64
}

// GridActor is a wrapper around the device object.
type GridActor struct {
	device    Device
	scheduler Scheduler
	inbox     chan interface{}
	state     ActorProcessState
}

// UpdateStatus requests actor to perform a device status read
type UpdateStatus struct{}

// WriteControl requests actor to perform a device control write
type WriteControl struct{}

// Start the actor
type Start struct{}

// Stop the actor
type Stop struct{}

// Kill the actor
type Kill struct{}

// NewGridActor is a factory function for GridActors
func NewGridActor(d Device) GridActor {
	return GridActor{device: d, inbox: nil, state: uninitialized}
}

// Initialize fulfills Initialize() in the AssetProcess Interface. The Actor is initilized and put in the Run state.
func (p *GridActor) Initialize() error {

	p.inbox = make(chan interface{})
	p.state = initialized

	p.scheduler = NewScheduler()
	read := NewTarget(p.inbox, UpdateStatus{}, 1000)
	write := NewTarget(p.inbox, WriteControl{}, 1000)
	p.scheduler.AddTarget(read)
	p.scheduler.AddTarget(write)

	go p.Run()
	return nil
}

// Run fulfills the AssetProcess Interface. In this state the Actor starts the scheduler and recieves messages.
func (p *GridActor) Run() error {
	p.scheduler.Run()
	p.state = running
	for {
		msg := <-p.inbox
		switch msg {
		case UpdateStatus{}:
			p.device.ReadDeviceStatus()
		case WriteControl{}:
			p.device.WriteDeviceControl()
		case Stop{}:
			go p.Stop()
			return nil
		case Kill{}:
			return nil
		}
	}
}

// Stop fulfills the AssetProcess Interface. In this state the Actor pauses the scheduler.
func (p *GridActor) Stop() error {
	p.scheduler.Stop()
	p.state = stopped
	for {
		msg := <-p.inbox
		switch msg {
		case Start{}:
			go p.Run()
			return nil
		case Kill{}:
			return nil
		}
	}
}
