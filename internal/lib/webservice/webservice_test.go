package webservice

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/lib/webservice/models"
	"gotest.tools/assert"
)

func TestAssetStatusGet(t *testing.T) {
	db, err := models.NewDB()
	assert.NilError(t, err)

	app := App{
		DB: db,
	}

	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	asset := models.AssetStatus{
		PID:  pid,
		Name: "ESS",
		KW:   rand.Float64(),
		KVAR: rand.Float64(),
	}

	_, err = app.DB.Exec(`INSERT INTO asset_status (pid, name, kw, kvar) 
			      VALUES ($1, $2, $3, $4)`,
		asset.PID, asset.Name, asset.KW, asset.KVAR)
	assert.NilError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/assets/"+asset.PID.String()+"/status", nil)

	router := app.Router()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "get returned 200")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")
}

func TestAssetStatusPost(t *testing.T) {
	db, err := models.NewDB()
	assert.NilError(t, err)

	app := App{
		DB: db,
	}

	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	asset := models.AssetStatus{
		PID:  pid,
		Name: "Grid",
		KW:   rand.Float64(),
		KVAR: rand.Float64(),
	}

	reqBody, err := json.Marshal(asset)

	var statusReturn models.AssetStatus
	err = json.Unmarshal(reqBody, &statusReturn)
	assert.NilError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "http://example.com/assets/"+asset.PID.String()+"/status", bytes.NewBuffer(reqBody))

	router := app.Router()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code, "post returned 201")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")

}

/*
func TestAssetControlGet(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	app := App{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/asset/"+pid.String()+"/control", nil)

	router := app.Router()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "get returned 200")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")
}
*/
