package dispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/mock"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type DummyDispatch struct {
	mux          *sync.Mutex
	PID          uuid.UUID
	AssetStatus  map[uuid.UUID]msg.Msg
	AssetControl map[uuid.UUID]interface{}
	msgList      []msg.Msg
}

func (d *DummyDispatch) UpdateStatus(msg msg.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.msgList = append(d.msgList, msg)
	d.AssetStatus[msg.PID()] = msg
}

func (d *DummyDispatch) DropAsset(uuid.UUID) error {
	return nil
}

func (d *DummyDispatch) GetControl(pid uuid.UUID) (interface{}, bool) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.AssetControl[pid] = mock.AssertedControl()
	return d.AssetControl[pid], true
}

func NewDummyDispatch() Dispatcher {
	status := make(map[uuid.UUID]msg.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	return &DummyDispatch{&sync.Mutex{}, pid, status, control, []msg.Msg{}}
}

func (d DummyDispatch) MsgList() []msg.Msg {
	return d.msgList
}
