package root

import (
	"testing"

	"github.com/ohowland/cgc_core/internal/pkg/bus"
	"github.com/ohowland/cgc_core/internal/pkg/dispatch/mockdispatch"
	"gotest.tools/assert"
)

func TestNewRootSystem(t *testing.T) {
	bg, err := bus.NewBusGraph()
	assert.NilError(t, err)

	_, err = NewSystem(&bg, mockdispatch.NewMockDispatch())
	assert.NilError(t, err)
}
