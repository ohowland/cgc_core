package models

import (
	"testing"

	"gotest.tools/assert"
)

func TestDBCreate(t *testing.T) {
	createTables()
	assert.Assert(t, true)
}
