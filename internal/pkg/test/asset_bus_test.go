package cgcintegrationtest

import (
	"testing"
	"time"

	"github.com/ohowland/cgc/internal/pkg/bus/acbus"

	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/ess/virtualess"
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

	bus1.AddMember(&ess1)

	relay1 := bus1.Relayer().(*virtualacbus.VirtualACBus)       // How to make ths interface explicit?
	device1 := ess1.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?
	device2 := ess2.DeviceController().(*virtualess.VirtualESS) // How to make this interface expicit?

	relay1.AddMember(device1)
	relay1.AddMember(device2)

	go func(*ess.Asset, *ess.Asset, *acbus.ACBus) {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			ess1.UpdateStatus()
			ess2.UpdateStatus()
			bus1.UpdateRelayer()
		}
	}(&ess1, &ess2, &bus1)

	ess1.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: true})
	ess2.WriteControl(ess.MachineControl{Run: true, KW: 0.0, KVAR: 0.0, Gridform: false})

	time.Sleep(10 * time.Second)

}
