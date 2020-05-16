package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/lib/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/lib/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/lib/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/lib/bus/ac/virtualacbus"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus/ac"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

/*
func buildDispatch() (dispatch.Dispatcher, error) {
	dispatch, err := manualdispatch.New("./config/dispatch/manualdispatch.json")
	return &dispatch, err
}
*/

func buildBus(dispatch dispatch.Dispatcher) (ac.Bus, error) {
	vrBus, err := virtualacbus.New("./config/bus/virtualACBus.json")
	return vrBus, err
}

/*
func buildBusGraph(buses map[uuid.UUID]bus.Bus) bus.BusGraph {
	return bus.NewBusGraph(buses)
}
*/

func buildAssets(bus *ac.Bus) (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)
	vrBus := bus.Relayer().(*virtualacbus.VirtualACBus)

	grid, err := virtualgrid.New("./config/asset/virtualGrid.json")
	if err != nil {
		panic(err)
	}

	assets[grid.PID()] = &grid
	bus.AddMember(&grid)

	vrGrid := grid.DeviceController().(*virtualgrid.VirtualGrid)
	vrBus.AddMember(vrGrid)

	ess, err := virtualess.New("./config/asset/virtualESS.json")
	if err != nil {
		panic(err)
	}

	assets[ess.PID()] = &ess
	bus.AddMember(&ess)

	vrEss := ess.DeviceController().(*virtualess.VirtualESS)
	vrBus.AddMember(vrEss)

	feeder, err := virtualfeeder.New("./config/asset/virtualFeeder.json")
	if err != nil {
		panic(err)
	}

	assets[feeder.PID()] = &feeder
	bus.AddMember(&feeder)

	vrFeeder := feeder.DeviceController().(*virtualfeeder.VirtualFeeder)
	vrBus.AddMember(vrFeeder)

	return assets, err
}

func launchUpdateLoop(assets map[uuid.UUID]asset.Asset) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		<-ticker.C
		for _, asset := range assets {
			asset.UpdateStatus()
		}
	}
}

func main() {
	log.Println("Starting CGC v0.1")

	log.Println("[MAIN] Building Dispatch")
	dispatch, err := buildDispatch()
	if err != nil {
		panic(err)
	}

	log.Println("[MAIN] Building Buses")
	bus, err := buildBuses(dispatch)
	if err != nil {
		panic(err)
	}

	/*
		log.Println("Assembling Bus Graph")
		busGraph, err := buildBusGraph(buses)
		if err != nil {
			panic(err)
		}
	*/

	log.Println("[MAIN] Building Assets")
	assets, err := buildAssets(&bus)
	if err != nil {
		panic(err)
	}

	/*
		log.Println("[Assembling Microgrid]")
		microgrid, err := buildMicrogrid(busGraph)
		if err != nil {
			panic(err)
		}
	*/

	/*
		microgrid.linkDispatch(dispatch)
		if err != nil {
			panic(err)
		}
	*/

	log.Println("[MAIN] Starting update loops")
	launchUpdateLoop(assets)

	/*
		log.Println("Starting Datalogging")
		launchDatalogging(assets)
	*/

	/*
		log.Println("Starting webserver")
		launchServer(assets)
	*/

	log.Println("[MAIN] Stopping system")
}
