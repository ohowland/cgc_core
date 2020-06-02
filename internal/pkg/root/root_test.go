package root

import (
	"testing"

	"github.com/ohowland/cgc/internal/pkg/bus"
	"gotest.tools/assert"
)

func TestNewRootSystem(t *testing.T) {
	bg, err := bus.NewBusGraph()
	assert.NilError(err)

	s, err := NewSystem()
	assert.NilError(err)
}
