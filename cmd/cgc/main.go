package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	"github.com/ohowland/cgc_core/internal/pkg/msg"
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
		bus = v.(*ac.Bus)
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

	log.Println("[Main] Starting update loops")
	busGraph.DumpString()
	var wg sync.WaitGroup
	wg.Add(2)
	go launchUpdateAssets(assets, sigs1, &wg)
	go monitorSystem(&system, sigs2, &wg)
	wg.Wait()

	/*
		log.Println("Starting Datalogging")
		launchDatalogging(assets)
	*/

	/*
		log.Println("Starting webserver")
		launchServer(assets)
	*/

	log.Println("[Main] Stopping system")
}

func monitorSystem(s *root.System, sigs chan os.Signal, wg *sync.WaitGroup) {
	ch, err := s.Subscribe(uuid.UUID{}, msg.Status)
	if err != nil {
		panic(err)
	}

	assets := make(map[uuid.UUID]interface{})
	staticOrder := make([]uuid.UUID, 0)
	ticker := time.NewTicker(1 * time.Second)
loop:
	for {
		select {
		case <-ticker.C:
			b := bufio.NewWriter(os.Stdout)
			printClearTerm()
			printTimestap(b)
			for _, pid := range staticOrder {
				p, _ := pid.MarshalText()
				s := []byte(fmt.Sprintf("%v", assets[pid]))
				printAssetStatus(b, p, s)
			}
			b.Flush()
		case msg := <-ch:
			if _, ok := assets[msg.PID()]; !ok {
				staticOrder = append(staticOrder, msg.PID())
			}
			assets[msg.PID()] = msg.Payload()
		case <-sigs:
			break loop
		}
	}
	log.Println("[System] Goroutine Shutdown")
	wg.Done()
}

func printClearTerm() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printTimestap(b *bufio.Writer) {
	t := time.Now()
	s, _ := t.MarshalText()
	b.Write(s)
	b.WriteString("\n")
}

func printAssetStatus(b *bufio.Writer, pid []byte, status []byte) {
	b.Write(pid)
	b.WriteString(": ")
	b.Write(status)
	b.WriteString("\n")
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
