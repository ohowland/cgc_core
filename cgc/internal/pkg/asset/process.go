package asset

type AssetProcess interface {
	Initialize() error
	Run() error
	Stop() error
}

type ProcessState int

const (
	uninitialized ProcessState = iota
	initialized   ProcessState = iota
	running       ProcessState = iota
	stopped       ProcessState = iota
)
