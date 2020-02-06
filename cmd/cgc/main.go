package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/bus"
)

func buildBuses() (map[uuid.UUID]bus.Bus, error) {
	buses := make(map[uuid.UUID]bus.Bus)
	return buses, nil
}

/*
func buildBusGraph(buses map[uuid.UUID]bus.Bus) bus.BusGraph {
	return bus.NewBusGraph(buses)
}
*/

func buildAssets(buses map[uuid.UUID]bus.Bus) (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	return assets, nil
}

func launchUpdateLoop(assets map[uuid.UUID]asset.Asset) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		<-ticker.C
		for _, asset := range assets {
			asset.UpdateStatus()
			//asset.WriteControl()
		}
		break
	}
	return nil
}

/*
func launchServer(assets map[uuid.UUID]asset.Asset) {
	go web.StartServer(assets)
}
*/

func main() {
	log.Println("Starting CGC v0.1")

	log.Println("Building Buses")
	buses, err := buildBuses()
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

	log.Println("[Building Assets]")
	assets, err := buildAssets(buses)
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
		log.Println("[Starting Dispatch]")
		dispatch, err := buildDispatch(busGraph)
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

	log.Println("Starting update loops")
	launchUpdateLoop(assets)
	/*
		log.Println("Starting dispatch loop")
		launchDispatchLoop(dispatch)
	*/

	/*
		log.Println("Starting Datalogging")
		launchDatalogging(assets)
	*/

	/*
		log.Println("Starting webserver")
		launchServer(assets)
	*/

	log.Println("Stopping system")
}
