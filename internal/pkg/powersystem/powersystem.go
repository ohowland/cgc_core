import (
	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/bus"
)

type Model struct {
	root       bus.Bus
	buses      map[string]bus.Bus
	composites map[uuid.UUID]asset.Composite
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