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
	"github.com/ohowland/cgc/internal/pkg/bus/virtualacbus"
)

func main() {
	log.Println("[starting]")
	log.Println("[loading assets]")
	assets, err := loadAssets()
	if err != nil {
		panic(err)
	}

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

	log.Println("[stopping]")
	os.Exit(0)
}

func loadAssets() (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	bus, err := virtualacbus.New("../../config/bus/virtualACBus.json")

	grid, err := virtualgrid.New("../../config/asset/virtualGrid.json", bus)
	if err != nil {
		return assets, err
	}
	assets[grid.PID()] = &grid

	ess, err := virtualess.New("../../config/asset/virtualESS.json", bus)
	if err != nil {
		return assets, err
	}
	assets[ess.PID()] = &ess

	pv, err := virtualpv.New("../../config/asset/virtualPV.json")
	if err != nil {
		return assets, err
	}
	assets[pv.PID()] = &pv

	feeder, err := virtualfeeder.New("../../config/asset/virtualFeeder.json", bus)
	if err != nil {
		return assets, err
	}
	assets[feeder.PID()] = &feeder

	return assets, nil
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
