package cgc

// EssAsset is a datastructure for an Energy Storage System Asset
type EssAsset struct {
	status EssStatus
}

// NewEssAsset returns an initalized EssAsset
func NewEssAsset() EssAsset {
	return EssAsset{
		status: EssStatus{kw: 1.0, kvar: 2.0, kwh: 3.0},
	}
}

func (a EssAsset) kw() float64 {
	return a.status.kw
}

func (a EssAsset) kvar() float64 {
	return a.status.kvar
}

// EssStatus is a data structure representing an architypical energy storage system
type EssStatus struct {
	kw   float64
	kvar float64
	kwh  float64
}
