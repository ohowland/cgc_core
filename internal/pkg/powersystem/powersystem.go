import (
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

type Microgrid struct {
	model    bus.BusGraph
	dispatch dispatch.Dispatcher
}

func newPowerSystem(configPath) Microgrid {
	buses, err := buildBuses()
	if err != nil {
		panic(err)
	}

	assets, err := buildAssets(buses)
	if err != nil {
		panic(err)
	}

	composites, err := buildComposites(assets)
	if err != nil {
		panic(err)
	}

	return Microgrid{}
}