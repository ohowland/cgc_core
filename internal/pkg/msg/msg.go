package msg

import "github.com/google/uuid"

// Publisher is an interface for objects that allow subscribtion to their events
type Publisher interface {
	Subscribe(uuid.UUID) <-chan Msg
	Unsubscribe(uuid.UUID)
}

// Msg is
type Msg struct {
	sender  uuid.UUID
	payload interface{}
}

// New is the Msg factor function
func New(sender uuid.UUID, payload interface{}) Msg {
	return Msg{sender, payload}
}

// PID returns the sender's PID
func (v Msg) PID() uuid.UUID {
	return v.sender
}

// Payload returns the message data
func (v Msg) Payload() interface{} {
	return v.payload
}
