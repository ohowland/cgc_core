package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
	"github.com/ohowland/cgc/internal/pkg/web"
)

func buildBuses() (map[uuid.UUID]bus.Bus, error) {
	buses := make(map[uuid.UUID]bus.Bus)

	bus, err := virtualacbus.New("./config/bus/virtualACBus.json")
	if err != nil {
		return buses, err
	}
	buses[bus.PID()] = &bus

	return buses, err
}

func buildBusGraph(buses map[uuid.UUID]bus.Bus) bus.BusGraph {
	return bus.NewBusGraph(buses)
}

func buildAssets(buses map[string]bus.Bus) (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	ess, err := virtualess.New("./config/asset/virtualESS.json")
	if err != nil {
		return assets, err
	}
	assets[ess.PID()] = &ess

	grid, err := virtualgrid.New("./config/asset/virtualGrid.json", buses)
	if err != nil {
		return assets, err
	}
	assets[grid.PID()] = &grid
	/*
		pv, err := virtualpv.New("./config/asset/virtualPV.json", buses)
		if err != nil {
			return assets, err
		}
		assets[pv.PID()] = &pv
	*/

	feeder, err := virtualfeeder.New("./config/asset/virtualFeeder.json", buses)
	if err != nil {
		return assets, err
	}
	assets[feeder.PID()] = &feeder

	return assets, nil
}

func launchUpdateLoop(assets map[uuid.UUID]asset.Asset) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		<-ticker.C
		for _, asset := range assets {
			asset.UpdateStatus()
			asset.WriteControl()
		}
	}
	return nil
}

func launchServer(assets map[uuid.UUID]asset.Asset) {
	go web.StartServer(assets)
}

func main() {
	log.Println("Starting CGC v0.1")

	log.Println("Building Buses")
	buses, err := buildBuses()
	if err != nil {
		panic(err)
	}

	log.Println("Assembling Bus Graph")
	busGraph, err := buildBusGraph(buses)
	if err != nil {
		panic(err)
	}

	log.Println("[Building Assets]")
	assets, err := buildAssets(buses)
	if err != nil {
		panic(err)
	}

	log.Println("[Assembling Microgrid]")
	microgrid, err := buildMicrogrid(busGraph)
	if err != nil {
		panic(err)
	}

	log.Println("[Starting Dispatch]")
	dispatch, err := buildDispatch(busGraph)
	if err != nil {
		panic(err)
	}

	microgrid.linkDispatch(dispatch)
	if err != nil {
		panic(err)
	}

	log.Println("Starting update loops")
	launchUpdateLoop(assets)

	log.Println("Starting dispatch loop")
	launchDispatchLoop(dispatch)

	log.Println("Starting Datalogging")
	launchDatalogging(assets)

	log.Println("Starting webserver")
	launchServer(assets)

	log.Println("Stopping system")
}
