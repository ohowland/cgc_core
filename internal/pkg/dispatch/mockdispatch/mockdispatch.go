package mockdispatch

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset/mockasset"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

type MockDispatch struct {
	mux          *sync.Mutex
	pid          uuid.UUID
	pub          msg.Publisher
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
	d.AssetControl[pid] = mockasset.AssertedControl()
	return d.AssetControl[pid], true
}

func NewMockDispatch() dispatch.Dispatcher {
	status := make(map[uuid.UUID]msg.Msg)
	control := make(map[uuid.UUID]interface{})
	pid, _ := uuid.NewUUID()
	pub := msg.NewPublisher(pid)
	return &MockDispatch{&sync.Mutex{}, pid, pub, status, control, []msg.Msg{}}
}

func (d MockDispatch) MsgList() []msg.Msg {
	return d.msgList
}

func (d MockDispatch) Subscribe(pid uuid.UUID, topic msg.Topic) (<-chan msg.Msg, error) {
	return d.pub.Subscribe(pid, topic)
}
func (d MockDispatch) Unsubscribe(pid uuid.UUID) {}

func (d MockDispatch) PID() uuid.UUID {
	return d.pid
}

func (d MockDispatch) StartProcess(<-chan msg.Msg) error {
	return nil
}
