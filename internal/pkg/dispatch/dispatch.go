package dispatch

import "github.com/google/uuid"

type Dispatcher interface {
	UpdateStatus(uuid.UUID, interface{})
	DropStatus(uuid.UUID)
	GetControl() map[uuid.UUID]interface{}
}
