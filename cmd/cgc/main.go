package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/bus/virtualbus"
	"github.com/ohowland/cgc/internal/pkg/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid/virtualgrid"
)

func main() {
	log.Println("[starting]")
	log.Println("[loading assets]")
	assets, err := loadAssets()
	if err != nil {
		panic(err)
	}

	log.Println("[launching processes]")
	processes, err := launchAssets(assets)
	if err != nil {
		panic(err)
	}

	log.Println("[running]")
	time.Sleep(time.Duration(10) * time.Second)

	log.Println("[stopping]")
	stopAssets(processes)
	os.Exit(0)
}

func loadAssets() (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	grid, err := virtualgrid.New("../../config/asset/virtualGrid.json")
	if err != nil {
		return assets, err
	}
	assets[grid.PID()] = &grid

	ess, err := virtualess.New("../../config/asset/virtualESS.json")
	if err != nil {
		return assets, err
	}
	assets[ess.PID()] = &ess

	feeder, err := virtualfeeder.New("../../config/asset/virtualFeeder.json")
	if err != nil {
		return assets, err
	}
	assets[feeder.PID()] = &feeder

	bus, err := virtualbus.New("../../config/asset/virtualBus.json")
	if err != nil {
		return assets, err
	}

	for _, asset := range assets {
		bus.Link(asset)
	}

	assets[bus.PID()] = &bus

	/*
		pv, err := virtualpv.New("../../config/asset/virtualPV.json")
		if err != nil {
			return assets, err
		}
		assets[pv.PID()] = &pv
	*/

	go vsm.RunVirtualSystem()

	return assets, nil
}

func launchAssets(assets map[uuid.UUID]asset.Asset) (map[uuid.UUID]chan interface{}, error) {
	inboxes := make(map[uuid.UUID]chan interface{})
	for _, a := range assets {
		inboxes[a.PID()] = asset.StartProcess(a)
	}

	return inboxes, nil
}

func startAssets(assets map[uuid.UUID]chan interface{}) {
	for _, inbox := range assets {
		inbox <- asset.Start{}
	}
}

func stopAssets(assets map[uuid.UUID]chan interface{}) {
	for _, inbox := range assets {
		inbox <- asset.Stop{}
	}
}

type systemConfig struct {
	AssetPaths []string `json:"AssetPaths"`
}

func readSystemConfig(path string) (systemConfig, error) {
	c := systemConfig{}
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(jsonFile, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}
