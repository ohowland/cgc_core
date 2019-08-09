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
