package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/lib/asset/ess/virtualess"
	"github.com/ohowland/cgc_core/internal/lib/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc_core/internal/lib/asset/grid/virtualgrid"
	"github.com/ohowland/cgc_core/internal/lib/bus/ac/virtualacbus"
	"github.com/ohowland/cgc_core/internal/pkg/asset"
	"github.com/ohowland/cgc_core/internal/pkg/asset/ess"
	"github.com/ohowland/cgc_core/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc_core/internal/pkg/asset/grid"
	"github.com/ohowland/cgc_core/internal/pkg/bus"
	"github.com/ohowland/cgc_core/internal/pkg/bus/ac"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch/manualdispatch"
	"github.com/ohowland/cgc_core/internal/pkg/mongodb"
	"github.com/ohowland/cgc_core/internal/pkg/root"
)

func main() {
	log.Println("[Main] Starting CGC_Core v0.0.1")
	sigs1 := make(chan os.Signal, 1)
	sigs2 := make(chan os.Signal, 1)
	signal.Notify(sigs1, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigs2, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[Main] Building Buses")
	buses, err := buildBuses()
	if err != nil {
		panic(err)
	}

	var bus *ac.Bus
	for _, v := range buses {
		bus = v.(*ac.Bus) // random root bus
		break
	}

	log.Println("[Main] Building Assets")
	assets, err := buildAssets(bus)
	if err != nil {
		panic(err)
	}

	log.Println("[Main] Assembling Bus Graph")
	busGraph, err := buildBusGraph(bus, buses, assets)
	if err != nil {
		panic(err)
	}

	log.Println("[Main] Building Dispatcher")
	dispatch, err := buildDispatch()
	if err != nil {
		panic(err)
	}

	log.Println("[Main] Assembling System")
	system, err := buildSystem(&busGraph, dispatch)
	if err != nil {
		panic(err)
	}

	log.Println("[Main] Connecting MongoDB Service")
	err = linkWebservice(&system)
	if err != nil {
		panic(err)
	}

	log.Println("[Main] Propigate Configuration")
	propigateConfigurations(buses, assets)
	if err != nil {
		panic(err)
	}

	log.Println("[Main] Starting update loops")
	var wg sync.WaitGroup
	wg.Add(1)
	go launchUpdateAssets(assets, sigs1, &wg)
	wg.Wait()

	log.Println("[Main] Stopping system")
}

func launchUpdateAssets(assets map[uuid.UUID]asset.Asset, sigs chan os.Signal, wg *sync.WaitGroup) {
	ticker := time.NewTicker(100 * time.Millisecond)
loop:
	for {
		select {
		case <-ticker.C:
			for _, asset := range assets {
				asset.UpdateStatus()
			}
		case <-sigs:
			for _, asset := range assets {
				go asset.Shutdown(wg)
			}
			time.Sleep(1 * time.Second)
			break loop
		}
	}
	log.Println("[UpdateAssets] Goroutine Shutdown")
	wg.Done()
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

func buildBuses() (map[uuid.UUID]bus.Bus, error) {
	buses := make(map[uuid.UUID]bus.Bus)
	vrBus1, err := virtualacbus.New("./config/bus/virtualACBus1.json")
	buses[vrBus1.PID()] = &vrBus1

	if err != nil {
		return buses, err
	}

	vrBus2, err := virtualacbus.New("./config/bus/virtualACBus2.json")
	buses[vrBus2.PID()] = &vrBus2

	return buses, err
}

func buildBusGraph(rootBus bus.Bus, buses map[uuid.UUID]bus.Bus, assets map[uuid.UUID]asset.Asset) (bus.BusGraph, error) {
	g, err := bus.BuildBusGraph(rootBus, buses, assets)
	return g, err
}

func buildDispatch() (dispatch.Dispatcher, error) {
	return manualdispatch.New("./config/dispatch/manualdispatch.json")
}

func buildSystem(g *bus.BusGraph, d dispatch.Dispatcher) (root.System, error) {
	return root.NewSystem(g, d)
}

func linkWebservice(sys *root.System) error {
	mongoHandler, err := mongodb.New("./config/database/mongodb_config.json", sys)
	go mongoHandler.Process()
	return err
}

func propigateConfigurations(buses map[uuid.UUID]bus.Bus, assets map[uuid.UUID]asset.Asset) {
	for _, bus := range buses {
		bus.UpdateConfig()
	}

	for _, asset := range assets {
		asset.UpdateConfig()
	}
}

/*
func monitorSystem(s *root.System, sigs chan os.Signal, wg *sync.WaitGroup) {
	statusCh, err := s.Subscribe(uuid.UUID{}, msg.Status)
	configCh, err := s.Subscribe(uuid.UUID{}, msg.Config)
	if err != nil {
		panic(err)
	}

	status := make(map[uuid.UUID]interface{})
	config := make(map[uuid.UUID]interface{})

	orderedStatus := make([]uuid.UUID, 0)
	ticker := time.NewTicker(1 * time.Second)
loop:
	for {
		select {
		case <-ticker.C:
			for _, pid := range orderedStatus {
				cfg, ok1 := config[pid]
				assetCfg, ok2 := cfg.(asset.Config)
				if ok1 && ok2 {
					log.Println(assetCfg.Name(), status[pid])
				} else {
					log.Println(pid, status[pid])
				}
			}
		case msg := <-statusCh:
			if _, ok := status[msg.PID()]; !ok {
				orderedStatus = append(orderedStatus, msg.PID())
			}
			status[msg.PID()] = msg.Payload()
		case msg := <-configCh:
			if _, ok := config[msg.PID()]; !ok {

			}
		case <-sigs:
			break loop
		}
	}
	log.Println("[System] Goroutine Shutdown")
	wg.Done()
}
*/
