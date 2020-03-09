package dispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

type DummyDispatch struct {
	mux          *sync.Mutex
	PID          uuid.UUID
	assetStatus  map[uuid.UUID]asset.Msg
	assetControl map[uuid.UUID]interface{}
}

func (d *DummyDispatch) UpdateStatus(msg asset.Msg) {
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

func NewDummyDispatch() Dispatcher {
	status := make(map[uuid.UUID]asset.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	return &DummyDispatch{&sync.Mutex{}, pid, status, control}
}
