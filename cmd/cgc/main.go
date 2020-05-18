package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/lib/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/lib/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/lib/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/lib/bus/ac/virtualacbus"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/ac"
)

/*
func buildDispatch() (dispatch.Dispatcher, error) {
	dispatch, err := manualdispatch.New("./config/dispatch/manualdispatch.json")
	return &dispatch, err
}
*/

/*
func buildBusGraph(buses map[uuid.UUID]bus.Bus) bus.BusGraph {
	return bus.NewBusGraph(buses)
}
*/

func main() {
	log.Println("Starting cgc v0.1.1")
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	/*
		log.Println("[MAIN] Building Dispatch")
		dispatch, err := buildDispatch()
		if err != nil {
			panic(err)
		}
	*/

	log.Println("[MAIN] Building Buses")
	bus, err := buildBus()
	if err != nil {
		panic(err)
	}

	log.Println("[MAIN] Building Assets")
	assets, err := buildAssets(&bus)
	if err != nil {
		panic(err)
	}

	log.Println("[MAIN] Assembling Bus Graph")
	busGraph, err := buildBusGraph(&bus, assets)
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
	launchUpdateLoop(&busGraph, assets, sigs)

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

func launchUpdateLoop(g *bus.BusGraph, assets map[uuid.UUID]asset.Asset, sigs chan os.Signal) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, asset := range assets {
				asset.UpdateStatus()
			}
		case <-sigs:
			var wg sync.WaitGroup
			for _, asset := range assets {
				go asset.Shutdown(&wg)
			}
			wg.Wait()
			time.Sleep(1 * time.Second)
			return
		}
	}
}

func buildAssets(bus *ac.Bus) (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	grid := buildVirtualGridAsset(bus)
	ess := buildVirtualESSAsset(bus)
	feeder := buildVirtualFeederAsset(bus)

	assets[ess.PID()] = ess
	assets[grid.PID()] = grid
	assets[feeder.PID()] = feeder

	return assets, nil
}

func buildVirtualGridAsset(bus *ac.Bus) *grid.Asset {
	grid, err := virtualgrid.New("./config/asset/virtualGrid.json")
	if err != nil {
		panic(err)
	}

	vrBus := bus.Relayer().(*virtualacbus.VirtualACBus)
	vrGrid := grid.DeviceController().(*virtualgrid.VirtualGrid)
	vrBus.AddMember(vrGrid)

	return &grid
}

func buildVirtualESSAsset(bus *ac.Bus) *ess.Asset {
	ess, err := virtualess.New("./config/asset/virtualESS.json")
	if err != nil {
		panic(err)
	}

	vrBus := bus.Relayer().(*virtualacbus.VirtualACBus)
	vrEss := ess.DeviceController().(*virtualess.VirtualESS)
	vrBus.AddMember(vrEss)

	return &ess
}

func buildVirtualFeederAsset(bus *ac.Bus) *feeder.Asset {
	feeder, err := virtualfeeder.New("./config/asset/virtualFeeder.json")
	if err != nil {
		panic(err)
	}

	vrBus := bus.Relayer().(*virtualacbus.VirtualACBus)
	vrFeeder := feeder.DeviceController().(*virtualfeeder.VirtualFeeder)
	vrBus.AddMember(vrFeeder)

	return &feeder
}

func buildBus() (ac.Bus, error) {
	vrBus, err := virtualacbus.New("./config/bus/virtualACBus.json")
	return vrBus, err
}

func buildBusGraph(rootBus bus.Bus, assets map[uuid.UUID]asset.Asset) (bus.BusGraph, error) {
	buses := make(map[uuid.UUID]bus.Bus)
	buses[rootBus.PID()] = rootBus
	g, err := bus.BuildBusGraph(rootBus, buses, assets)
	return g, err
}
