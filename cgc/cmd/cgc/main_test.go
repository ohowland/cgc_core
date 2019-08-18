package cgc

import (
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"gotest.tools/assert"
)

type DummyAsset struct {
	pid uuid.UUID
}
type DummyAssetStatus struct{}
type DummyAssetControl struct{}
type DummyAssetConfig struct{}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d DummyAsset) Status() interface{} {
	return DummyAssetStatus{}
}

func (d DummyAsset) Control(interface{}) {

}

func (d DummyAsset) Config() interface{} {
	return DummyAssetConfig{}
}

func (d DummyAsset) UpdateStatus() error {
	log.Print("reading dummy device...")
	time.Sleep(time.Duration(100) * time.Millisecond)
	return nil
}

func (d DummyAsset) WriteControl() error {
	log.Print("writing dummy device...")
	time.Sleep(time.Duration(100) * time.Millisecond)
	return nil
}

func TestLaunch(t *testing.T) {
	asts := make(map[uuid.UUID]asset.Asset)
	pid, err := uuid.NewUUID()
	d := DummyAsset{pid}
	asts[d.pid] = d
	inboxes, err := launchAssets(asts)

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Duration(2) * time.Second)
	stopAssets(inboxes)
}

func TestReadSystemConfig(t *testing.T) {

	assertConfig := []string{
		"../../internal/pkg/sel1547/sel1547",
		"../../internal/pkg/sel1547/sel1547",
	}
	config, err := readSystemConfig("./settings_test.json")

	if err != nil {
		t.Fatal(err)
	}

	for i := range assertConfig {
		assert.Assert(t, config.AssetPaths[i] == assertConfig[i])
	}
}
