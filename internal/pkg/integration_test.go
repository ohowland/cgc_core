package cgcintegrationtest

import (
	"log"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"
	"github.com/ohowland/cgc/internal/pkg/dispatch/manualdispatch"
	"github.com/ohowland/cgc/internal/pkg/msg"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus/virtualacbus"
)

func TestVirtualBusVirtualEss(t *testing.T) {
	disp, err := manualdispatch.New("")

	bus1, err := virtualacbus.New("../../config/bus/virtualACBus.json", &disp)
	if err != nil {
		t.Fatal(err)
	}

	ess1, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess2, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess3, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess4, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess5, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess6, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	bus1.AddMember(&ess1)
	bus1.AddMember(&ess2)
	bus1.AddMember(&ess3)
	bus1.AddMember(&ess4)
	bus1.AddMember(&ess5)
	bus1.AddMember(&ess6)

	relay1 := bus1.Relayer().(*virtualacbus.VirtualACBus)       // How to make ths interface explicit?
	device1 := ess1.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?
	device2 := ess2.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?
	device3 := ess3.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?
	device4 := ess4.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?
	device5 := ess5.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?
	device6 := ess6.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?

	relay1.AddMember(device1)
	relay1.AddMember(device2)
	relay1.AddMember(device3)
	relay1.AddMember(device4)
	relay1.AddMember(device5)
	relay1.AddMember(device6)

	go func(*ess.Asset, *ess.Asset, *ess.Asset, *ess.Asset, *ess.Asset, *ess.Asset, *acbus.ACBus) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			go ess1.UpdateStatus()
			go ess2.UpdateStatus()
			go ess3.UpdateStatus()
			go ess4.UpdateStatus()
			go ess5.UpdateStatus()
			go ess6.UpdateStatus()
			//go bus1.UpdateRelayer()
		}
	}(&ess1, &ess2, &ess3, &ess4, &ess5, &ess6, &bus1)

	wPID, _ := uuid.NewUUID()
	writer1 := make(chan msg.Msg)
	writer2 := make(chan msg.Msg)
	writer3 := make(chan msg.Msg)
	writer4 := make(chan msg.Msg)
	writer5 := make(chan msg.Msg)
	writer6 := make(chan msg.Msg)

	ess1.RequestControl(wPID, writer1)
	ess2.RequestControl(wPID, writer2)
	ess3.RequestControl(wPID, writer3)
	ess4.RequestControl(wPID, writer4)
	ess5.RequestControl(wPID, writer5)
	ess6.RequestControl(wPID, writer6)

	writer1 <- msg.New(wPID, ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: true})
	writer2 <- msg.New(wPID, ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	writer3 <- msg.New(wPID, ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	writer4 <- msg.New(wPID, ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	writer5 <- msg.New(wPID, ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	writer6 <- msg.New(wPID, ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})

	time.Sleep(5 * time.Second)

	assert.Assert(t, bus1.Energized() == true)

	device1.StopProcess()
	device2.StopProcess()
	device3.StopProcess()
	device4.StopProcess()
	device5.StopProcess()
	device6.StopProcess()
	time.Sleep(500 * time.Millisecond)
}

func TestVirtualBusAllAssets(t *testing.T) {
	disp, err := manualdispatch.New("")

	bus1, err := virtualacbus.New("../../config/bus/virtualACBus.json", &disp)
	if err != nil {
		t.Fatal(err)
	}

	ess1, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	feeder1, err := virtualfeeder.New("../../config/asset/virtualFeeder.json")
	if err != nil {
		t.Fatal(err)
	}

	grid1, err := virtualgrid.New("../../config/asset/virtualGrid.json")
	if err != nil {
		t.Fatal(err)
	}

	bus1.AddMember(&ess1)
	bus1.AddMember(&feeder1)
	bus1.AddMember(&grid1)

	relay1 := bus1.Relayer().(*virtualacbus.VirtualACBus)                // How to make this interface explicit?
	device1 := ess1.DeviceController().(*virtualess.VirtualESS)          // How to make this interface explicit?
	device2 := feeder1.DeviceController().(*virtualfeeder.VirtualFeeder) // How to make this interface explicit?
	device3 := grid1.DeviceController().(*virtualgrid.VirtualGrid)       // How to make this interface explicit?

	relay1.AddMember(device1)
	relay1.AddMember(device2)
	relay1.AddMember(device3)

	go func(*ess.Asset, *feeder.Asset, *grid.Asset, *acbus.ACBus) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			go grid1.UpdateStatus()
			go ess1.UpdateStatus()
			go feeder1.UpdateStatus()
		}
	}(&ess1, &feeder1, &grid1, &bus1)

	kwSp := 2.0

	wpid, _ := uuid.NewUUID()
	essWriter := make(chan msg.Msg)
	_ = ess1.RequestControl(wpid, essWriter)
	essWriter <- msg.New(wpid, ess.MachineControl{Run: true, KW: kwSp, KVAR: 0.0, Gridform: false})

	feederWriter := make(chan msg.Msg)
	_ = feeder1.RequestControl(wpid, feederWriter)
	feederWriter <- msg.New(wpid, feeder.MachineControl{CloseFeeder: true})

	gridWriter := make(chan msg.Msg)
	_ = grid1.RequestControl(wpid, gridWriter)
	gridWriter <- msg.New(wpid, grid.MachineControl{CloseIntertie: true})

	pid, _ := uuid.NewUUID()
	ch := grid1.Subscribe(pid)
	time.Sleep(5 * time.Second)
	gridstatus := <-ch

	assert.Assert(t, gridstatus.Payload().(grid.Status).KW() == -1*kwSp)
	assert.Assert(t, bus1.Energized() == true)

	device1.StopProcess()
	device2.StopProcess()
	device3.StopProcess()
	time.Sleep(500 * time.Millisecond)
}

type MockDispatch struct {
	msgList    []msg.Msg
	controlMap map[uuid.UUID]interface{}
}

func newMockDispatch() MockDispatch {
	msgList := make([]msg.Msg, 0)
	controlMap := make(map[uuid.UUID]interface{})
	return MockDispatch{msgList, controlMap}
}

func (d *MockDispatch) UpdateStatus(msg msg.Msg) {
	d.msgList = append(d.msgList, msg)
}

func (d *MockDispatch) DropAsset(uuid.UUID) error {
	return nil
}

func (d MockDispatch) GetControl() map[uuid.UUID]interface{} {
	return d.controlMap
}

func TestBusDispatchForwarding(t *testing.T) {

	dispatch := newMockDispatch()

	bus1, err := virtualacbus.New("../../config/bus/virtualACBus.json", &dispatch)
	if err != nil {
		t.Fatal(err)
	}

	ess1, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	bus1.AddMember(&ess1)

	// TODO: This should be hidden. When a virtual device is added to a virtual bus, it should add itself.
	relay1 := bus1.Relayer().(*virtualacbus.VirtualACBus)
	device1 := ess1.DeviceController().(*virtualess.VirtualESS)
	relay1.AddMember(device1)

	go func(*ess.Asset) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			go ess1.UpdateStatus()
		}
	}(&ess1)

	for len(dispatch.msgList) < 1 {
		time.Sleep(100 * time.Millisecond)
	}

	assert.Assert(t, dispatch.msgList[0].PID() == ess1.PID())
}

func TestDispatchCalculatedStatusAggregate(t *testing.T) {

	dispatch, err := manualdispatch.New("")
	if err != nil {
		t.Fatal(err)
	}

	bus1, err := virtualacbus.New("../../config/bus/virtualACBus.json", &dispatch)
	if err != nil {
		t.Fatal(err)
	}

	ess1, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	grid1, err := virtualgrid.New("../../config/asset/virtualGrid.json")
	if err != nil {
		t.Fatal(err)
	}

	bus1.AddMember(&ess1)
	bus1.AddMember(&grid1)

	// TODO: This should be hidden. When a virtual device is added to a virtual bus, it should add itself.
	relay1 := bus1.Relayer().(*virtualacbus.VirtualACBus)
	device1 := ess1.DeviceController().(*virtualess.VirtualESS)
	device2 := grid1.DeviceController().(*virtualgrid.VirtualGrid) // How to make this interface explicit?

	relay1.AddMember(device1)
	relay1.AddMember(device2)

	go func(*ess.Asset, *grid.Asset, *acbus.ACBus) {
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			<-ticker.C
			go grid1.UpdateStatus()
			go ess1.UpdateStatus()
		}
	}(&ess1, &grid1, &bus1)

	kwSp := 5.0

	wpid, _ := uuid.NewUUID()
	essWriter := make(chan msg.Msg)
	_ = ess1.RequestControl(wpid, essWriter)
	essWriter <- msg.New(wpid, ess.MachineControl{Run: true, KW: kwSp, KVAR: 0.0, Gridform: false})

	gridWriter := make(chan msg.Msg)
	_ = grid1.RequestControl(wpid, gridWriter)
	gridWriter <- msg.New(wpid, grid.MachineControl{CloseIntertie: true})

	time.Sleep(2000 * time.Millisecond)

	memberStatus := dispatch.GetStatus()
	log.Println(memberStatus)

	//assert.Assert(t, memberStatus[ess1.PID()].(asset.Status).KW() == kwSp)
}
