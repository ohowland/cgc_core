package asset

import "github.com/ohowland/cgc/internal/pkg/bus"

type VirtualAsset interface {
	LinkVirtualDevice(bus bus.Bus) error
	StartVirtualDevice()
	StopVirtualDevice()
}
