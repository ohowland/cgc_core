package target

import (
	"testing"
)

func TestReadStaticConfig(t *testing.T) {
	c := sel1547StaticConfig{}
	c.readConfig("config.json")
	t.Logf("what is it? %v", c)
}

func TestNewAsset(t *testing.T) {
	a, err := NewAsset("config.json")

	if err != nil {
		t.Logf("error: %v", err)
		t.FailNow()
	}
	t.Logf("what is it now? %v", a)
}
