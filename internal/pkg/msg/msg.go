package msg

import "github.com/google/uuid"

// Publisher is an interface for objects that allow subscribtion to their events
type Publisher interface {
	Subscribe(uuid.UUID) <-chan Msg
	Unsubscribe(uuid.UUID)
}

type Topic int

const (
	STATUS Topic = iota
	CONTROL
	CONFIG
	JSON
)

// Msg is
type Msg struct {
	sender  uuid.UUID
	topic   Topic
	payload interface{}
}

// New is the Msg factor function
func New(sender uuid.UUID, topic Topic, payload interface{}) Msg {
	return Msg{sender, topic, payload}
}

// PID returns the sender's PID
func (m Msg) PID() uuid.UUID {
	return m.sender
}

// Topic returns the message type
func (m Msg) Topic() Topic {
	return m.topic
}

// Payload returns the message data
func (m Msg) Payload() interface{} {
	return m.payload
}
