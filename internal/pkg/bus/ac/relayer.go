package ac

import "github.com/ohowland/cgc_core/internal/pkg/asset"

// Relayer is the interface for a bus relayer.
type Relayer interface {
	asset.Frequency
	asset.Voltage
}
