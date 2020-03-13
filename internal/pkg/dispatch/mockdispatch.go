package dispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type DummyDispatch struct {
	mux          *sync.Mutex
	PID          uuid.UUID
	AssetStatus  map[uuid.UUID]msg.Msg
	AssetControl map[uuid.UUID]interface{}
}

func (d *DummyDispatch) UpdateStatus(msg msg.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.AssetStatus[msg.PID()] = msg
}

func (d *DummyDispatch) DropAsset(uuid.UUID) error {
	return nil
}

func (d *DummyDispatch) GetControl() map[uuid.UUID]interface{} {
	d.mux.Lock()
	defer d.mux.Unlock()
	for _, Msg := range d.AssetStatus {
		d.AssetControl[Msg.PID()] = asset.AssertedControl()
	}
	return d.AssetControl
}

func NewDummyDispatch() Dispatcher {
	status := make(map[uuid.UUID]msg.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	return &DummyDispatch{&sync.Mutex{}, pid, status, control}
}
