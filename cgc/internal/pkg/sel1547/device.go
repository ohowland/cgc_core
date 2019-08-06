package asset

type Device interface {
	ReadDeviceStatus() error
	WriteDeviceControl() error
}
