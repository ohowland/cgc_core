package ac

// Relayer is the interface for a bus relayer.
type Relayer interface {
	Hz() float64
	Volt() float64
}
