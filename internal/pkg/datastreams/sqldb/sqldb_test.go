package sqldb

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/asset/mockasset"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"gotest.tools/v3/assert"
)

func newHandler() (Handler, error) {
	pid, _ := uuid.NewUUID()
	pub := msg.NewPublisher(pid)
	return New("./db_config_test.json", pub)

}

func TestGetConfig(t *testing.T) {
	h, err := newHandler()
	assert.NilError(t, err)

	assert.Equal(t, h.config.Port, 3306)
	assert.Equal(t, h.config.Server, "localhost")
}

func TestDatabaseConnection(t *testing.T) {
	h, _ := newHandler()
	db, err := h.DB()
	defer db.Close()

	assert.NilError(t, err)
}

func TestInitDatabase(t *testing.T) {
	h, _ := newHandler()
	db, _ := h.DB()
	defer db.Close()

	err := initDBTables(db)
	assert.NilError(t, err)
}

func TestAddRowDatabase(t *testing.T) {
	h, _ := newHandler()
	db, _ := h.DB()
	defer db.Close()

	mock := mockasset.New()
	ch, err := mock.Subscribe(h.PID(), msg.Status)
	assert.NilError(t, err)
	mock.UpdateStatus()

	msg := <-ch
	db := updateRow(mock.PID(), msg)
}
