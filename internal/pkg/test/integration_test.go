package cgcintegrationtest

import (
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"

	"github.com/ohowland/cgc/internal/lib/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/lib/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/lib/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/lib/bus/ac/virtualacbus"
	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/bus/ac"
)

func TestSimpleVirtualBus(t *testing.T) {
	bus1, err := virtualacbus.New("../../../config/bus/virtualACBus.json")
	assert.NilError(t, err)
	ess1, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

	bus1.AddMember(&ess1)

	// How to make ths interface explicit?
	// The virtual system is hidden behind the relay because the relay is the busses interface
	// to its physical data
	relay, ok := bus1.Relayer().(*virtualacbus.VirtualACBus)
	assert.Assert(t, ok)

	// How to make this interface expicit?
	device, ok := ess1.DeviceController().(*virtualess.VirtualESS)
	assert.Assert(t, ok)

	relay.AddMember(device)

	pid, _ := uuid.NewUUID()
	read, err := bus1.Subscribe(pid, msg.Status)

	write := make(chan msg.Msg)
	bus1.RequestControl(pid, write)

	go func(ess1 *ess.Asset) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				ess1.UpdateStatus()
			case msg := <-read:
				switch status := msg.Payload().(type) {
				case ess.Status:
					t.Log(status)
					assert.Assert(t, true)
				default:
					t.Log(status)
					assert.Assert(t, false)
				}
			}
		}
	}(&ess1)

	write <- msg.New(pid, msg.New(ess1.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: true}))

	time.Sleep(5 * time.Second)
}

func TestVirtualBusVirtualEss(t *testing.T) {
	bus1, err := virtualacbus.New("../../../config/bus/virtualACBus.json")
	assert.NilError(t, err)

	ess1, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

	ess2, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

	ess3, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

	ess4, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

	ess5, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

	ess6, err := virtualess.New("../../../config/asset/virtualESS.json")
	assert.NilError(t, err)

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

	go func(*ess.Asset, *ess.Asset, *ess.Asset, *ess.Asset, *ess.Asset, *ess.Asset, *ac.Bus) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			ess1.UpdateStatus()
			ess2.UpdateStatus()
			ess3.UpdateStatus()
			ess4.UpdateStatus()
			ess5.UpdateStatus()
			ess6.UpdateStatus()
			//go bus1.UpdateRelayer()
		}
	}(&ess1, &ess2, &ess3, &ess4, &ess5, &ess6, &bus1)

	//	This doesn't work. Dispatch needs to generate the control

	wPID, _ := uuid.NewUUID()
	writer := make(chan msg.Msg)

	bus1.RequestControl(wPID, writer)

	writer <- msg.New(wPID, msg.New(ess1.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: true}))
	writer <- msg.New(wPID, msg.New(ess2.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false}))
	writer <- msg.New(wPID, msg.New(ess3.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false}))
	writer <- msg.New(wPID, msg.New(ess4.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false}))
	writer <- msg.New(wPID, msg.New(ess5.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false}))
	writer <- msg.New(wPID, msg.New(ess6.PID(), ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false}))

	time.Sleep(1 * time.Second)

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
	bus1, err := virtualacbus.New("../../../config/bus/virtualACBus.json")
	if err != nil {
		t.Fatal(err)
	}

	ess1, err := virtualess.New("../../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	feeder1, err := virtualfeeder.New("../../../config/asset/virtualFeeder.json")
	if err != nil {
		t.Fatal(err)
	}

	grid1, err := virtualgrid.New("../../../config/asset/virtualGrid.json")
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

	go func(*ess.Asset, *feeder.Asset, *grid.Asset, *ac.Bus) {
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
	ch, _ := grid1.Subscribe(pid, msg.Status)
	time.Sleep(5 * time.Second)
	gridstatus := <-ch

	assert.Assert(t, gridstatus.Payload().(grid.Status).KW() == -1*kwSp)
	assert.Assert(t, bus1.Energized() == true)

	device1.StopProcess()
	device2.StopProcess()
	device3.StopProcess()
	time.Sleep(500 * time.Millisecond)
}

/*
func TestBusDispatchForwarding(t *testing.T) {

	bus1, err := virtualacbus.New("../../config/bus/virtualACBus.json")
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

	for len(mockDispatch.MsgList()) < 1 {
		time.Sleep(100 * time.Millisecond)
	}

	assert.Assert(t, mockDispatch.MsgList()[0].PID() == ess1.PID())
}
*/

/*
func TestDispatchCalculatedStatusAggregate(t *testing.T) {
	bus1, err := virtualacbus.New("../../config/bus/virtualACBus.json")
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

	go func(*ess.Asset, *grid.Asset, *ac.Bus) {
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

	//assert.Assert(t, memberStatus[ess1.PID()].(asset.Status).KW() == kwSp)
}
*/
