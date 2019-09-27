package virtual

type DummyVirtualAsset struct {
	status Status
}

type Status struct {
	KW          float64
	KVAR        float64
	Gridforming bool
}
