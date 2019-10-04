package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/ess/virtualess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder/virtualfeeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid/virtualgrid"
	"github.com/ohowland/cgc/internal/pkg/asset/pv/virtualpv"
	"github.com/ohowland/cgc/internal/pkg/bus"
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

func main() {
	log.Println("[starting]")
	log.Println("[loading assets]")
	busses, err := buildBusses()
	assets, err := buildAssets(busses)
	if err != nil {
		panic(err)
	}
	launchUpdateLoop(assets)

	//composites, err := buildComposites(assets)

	//launchDispatch(assets)
	//launchHMI(assets)

	log.Println("[stopping]")
	os.Exit(0)
}

func buildBuses() (map[string]bus.Bus, error) {
	buses := make(map[string]bus.Bus)
	bus, err := virtualacbus.New("./config/bus/virtualACBus.json")

	busses[bus.Name()] = &bus

	return busses, err
}

func buildAssets(buses map[string]bus.Bus) (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	grid, err := virtualgrid.New("./config/asset/virtualGrid.json", buses)
	if err != nil {
		return assets, err
	}
	assets[grid.PID()] = &grid

	ess, err := virtualess.New("./config/asset/virtualESS.json", buses)
	if err != nil {
		return assets, err
	}
	assets[ess.PID()] = &ess

	pv, err := virtualpv.New("./config/asset/virtualPV.json", buses)
	if err != nil {
		return assets, err
	}
	assets[pv.PID()] = &pv

	feeder, err := virtualfeeder.New("./config/asset/virtualFeeder.json", buses)
	if err != nil {
		return assets, err
	}
	assets[feeder.PID()] = &feeder

	return assets, nil
}

func launchUpdateLoop(assets map[uuid.UUID]asset.Asset) error {
	ticker := time.NewTicker(1 * time.Second)
	var i int
	for {
		<-ticker.C
		for _, asset := range assets {
			asset.UpdateStatus()
		}
		i++
		if i > 10 {
			break
		}
	}
	return nil
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
