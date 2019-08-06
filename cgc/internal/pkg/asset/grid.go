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

// State of the GridActor
type GridActor struct {
	asset GridAsset
	inbox chan string
	state ProcessState
}

func NewGridActor(asset GridAsset) GridActor {
	return GridActor{asset: asset, inbox: nil, state: uninitialized}
}

// Initialize fulfills Initialize() in the AssetProcess Interface
func (p *GridActor) Initialize() error {

	p.inbox = make(chan string)
	p.state = initialized
	return nil
}

type UpdateStatus struct{}

type WriteControl struct{}

type Quit struct{}

func (p *GridActor) Run() {
	var msg string
	for {
		msg = <-p.inbox
		switch msg {
		case "quit":
			return
		}
	}
}
