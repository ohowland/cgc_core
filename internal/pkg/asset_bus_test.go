package cgcintegrationtest

import (
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/pkg/asset/pv"
	"github.com/ohowland/cgc/internal/pkg/asset/pv/virtualpv"
	"github.com/ohowland/cgc/internal/pkg/bus/acbus/virtualacbus"
)

func TestVirtualBusVirtualEss(t *testing.T) {
	bus1, err := virtualacbus.New("../../../config/bus/virtualACBus.json")
	if err != nil {
		t.Fatal(err)
	}

	ess1, err := virtualess.New("../../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess2, err := virtualess.New("../../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess3, err := virtualess.New("../../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess4, err := virtualess.New("../../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess5, err := virtualess.New("../../../config/asset/virtualESS.json")
	if err != nil {
		t.Fatal(err)
	}

	ess6, err := virtualess.New("../../../config/asset/virtualESS.json")
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
			go bus1.UpdateRelayer()
		}
	}(&ess1, &ess2, &ess3, &ess4, &ess5, &ess6, &bus1)

	ess1.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: true})
	ess2.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	ess3.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	ess4.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	ess5.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})
	ess6.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})

	time.Sleep(5 * time.Second)

	assert.Assert(t, bus1.Energized() == true)

	device1.StopProcess()
	device2.StopProcess()
	device3.StopProcess()
	device4.StopProcess()
	device5.StopProcess()
	device6.StopProcess()
	time.Sleep(100 * time.Millisecond)
	bus1.StopProcess()
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

	pv1, err := virtualpv.New("../../../config/asset/virtualPV.json")
	if err != nil {
		t.Fatal(err)
	}

	bus1.AddMember(&ess1)
	bus1.AddMember(&feeder1)
	bus1.AddMember(&grid1)
	bus1.AddMember(&pv1)

	relay1 := bus1.Relayer().(*virtualacbus.VirtualACBus)                // How to make this interface explicit?
	device1 := ess1.DeviceController().(*virtualess.VirtualESS)          // How to make this interface explicit?
	device2 := feeder1.DeviceController().(*virtualfeeder.VirtualFeeder) // How to make this interface explicit?
	device3 := grid1.DeviceController().(*virtualgrid.VirtualGrid)       // How to make this interface explicit?
	device4 := pv1.DeviceController().(*virtualpv.VirtualPV)             // How to make this interface explicit?

	relay1.AddMember(device1)
	relay1.AddMember(device2)
	relay1.AddMember(device3)
	relay1.AddMember(device4)

	go func(*ess.Asset, *feeder.Asset, *grid.Asset, *pv.Asset, *acbus.ACBus) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			go ess1.UpdateStatus()
			go feeder1.UpdateStatus()
			go grid1.UpdateStatus()
			go pv1.UpdateStatus()
			go bus1.UpdateRelayer()
		}
	}(&ess1, &feeder1, &grid1, &pv1, &bus1)

	kwSp := 2.0

	ess1.WriteControl(ess.MachineControl{Run: true, KW: kwSp, KVAR: 0.0, Gridform: false})
	feeder1.WriteControl(feeder.MachineControl{CloseFeeder: true})
	grid1.WriteControl(grid.MachineControl{CloseIntertie: true})
	pv1.WriteControl(pv.MachineControl{Run: true, KWLimit: 10, KVAR: 0.0})

	pid, _ := uuid.NewUUID()
	ch := grid1.Subscribe(pid)
	time.Sleep(5 * time.Second)
	gridstatus := <-ch

	assert.Assert(t, gridstatus.KW() == -1*kwSp)
	assert.Assert(t, bus1.Energized() == true)

	device1.StopProcess()
	device2.StopProcess()
	device3.StopProcess()
	device4.StopProcess()
	time.Sleep(100 * time.Millisecond)
	bus1.StopProcess()
}