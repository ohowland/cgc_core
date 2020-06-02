package mockdispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset/mock"
	"github.com/ohowland/cgc/internal/pkg/msg"
)

type MockDispatch struct {
	mux          *sync.Mutex
	PID          uuid.UUID
	AssetStatus  map[uuid.UUID]msg.Msg
	AssetControl map[uuid.UUID]interface{}
	msgList      []msg.Msg
}

func (d *MockDispatch) UpdateStatus(msg msg.Msg) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.msgList = append(d.msgList, msg)
	d.AssetStatus[msg.PID()] = msg
}

func (d *MockDispatch) DropAsset(uuid.UUID) error {
	return nil
}

func (d *MockDispatch) GetControl(pid uuid.UUID) (interface{}, bool) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.AssetControl[pid] = mock.AssertedControl()
	return d.AssetControl[pid], true
}

func NewMockDispatch() Dispatcher {
	status := make(map[uuid.UUID]msg.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	return &MockDispatch{&sync.Mutex{}, pid, status, control, []msg.Msg{}}
}git 

func (d MockDispatch) MsgList() []msg.Msg {
	return d.msgList
}
