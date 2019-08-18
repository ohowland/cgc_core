package asset

// Asset is the interface for all physical devices that make up dispatchable sources/sinks in the power system.
type Asset interface {
	UpdateStatus() error
	WriteControl() error
	Status() interface{}
	Control(interface{})
	Config(interface{})
}
