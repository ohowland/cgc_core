package cgc

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/ohowland/cgc/internal/pkg/sel1547"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

func main() {

	// create assets and wrap in process.
	config, err := readSystemConfig("./settings.json")
	if err != nil {
		panic(err)
	}

	assets, err := loadAssets(config.AssetPaths)
	if err != nil {
		panic(err)
	}

	processes, err := launchAssets(assets)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Duration(1) * time.Second)
	stopAssets(processes)

}

func loadAssets(paths []string) (map[uuid.UUID]asset.Asset, error) {
	assets := make(map[uuid.UUID]asset.Asset)

	for _, path := range paths {
		// TODO: How to dynamically load in types
		// What if we don't know the type at compile time?
		a, err := sel1547.ConfigureAsset(path)
		if err != nil {
			return assets, err
		}
		assets[a.PID()] = a
	}
	return assets, nil
}

func launchAssets(assets map[uuid.UUID]asset.Asset) (map[uuid.UUID]chan interface{}, error) {
	inboxes := make(map[uuid.UUID]chan interface{})
	for _, a := range assets {
		inboxes[a.PID()] = asset.InitializeProcess(a.(asset.Device))
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
