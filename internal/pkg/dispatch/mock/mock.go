package mock

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type DummyDispatch struct {
	mux          *sync.Mutex
	PID          uuid.UUID
	assetStatus  map[uuid.UUID]msg.Msg
	assetControl map[uuid.UUID]interface{}
}

func (d *DummyDispatch) UpdateStatus(msg msg.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.assetStatus[msg.PID()] = msg
}

func (d *DummyDispatch) DropAsset(uuid.UUID) error {
	return nil
}

func (d *DummyDispatch) GetControl() map[uuid.UUID]interface{} {
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, Msg := range d.assetStatus {
		d.assetControl[Msg.PID()] = 0xBEEF
	}
	return d.assetControl
}

func NewDummyDispatch() dispatch.Dispatcher {
	status := make(map[uuid.UUID]msg.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	return &DummyDispatch{&sync.Mutex{}, pid, status, control}
}
