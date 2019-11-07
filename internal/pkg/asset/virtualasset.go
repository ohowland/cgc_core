package asset

import "github.com/ohowland/cgc/internal/pkg/bus"

// VirtualAsset defines the interface to the virtual assets.
type VirtualAsset interface {
	LinkToBus(bus bus.Bus) error
	StartVirtualDevice()
	StopVirtualDevice()
}
