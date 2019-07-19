package cgc

type stateIn struct {
	runRequest bool
}

type stateOut struct {
	running bool
}

type state interface {
	name() string             // name of state
	transition(stateIn) state // transition check function
	action() stateOut         // state action
}

type offState struct{}

func (offState) name() string {
	return "Off State"
}

func (offState) transition(in stateIn) state {
	if in.runRequest == true {
		return onState{}
	}
	return offState{}
}

func (offState) action() stateOut {
	return stateOut{running: false}
}

type onState struct{}

func (onState) name() string {
	return "On State"
}

func (onState) transition(in stateIn) state {
	if in.runRequest == false {
		return offState{}
	}
	return onState{}

}

func (onState) action() stateOut {
	return stateOut{running: true}
}

type statemachine struct {
	currentState state
}
