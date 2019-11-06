package asset

import "github.com/ohowland/cgc/internal/pkg/bus"

type VirtualAsset interface {
	LinkToBus(bus bus.Bus) error
	StartVirtualDevice()
	StopVirtualDevice()
}
