package webservice

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

func TestAssetStatusGet(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/asset/"+pid.String()+"/status", nil)

	router := makeRouter()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "get returned 200")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")

	status := AssetStatus{}
	err = json.Unmarshal(w.Body.Bytes(), &status)
	assert.NilError(t, err)
	assert.Equal(t, pid, status.PID)
}

func TestAssetStatusPost(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	reqBody, err := json.Marshal(AssetStatus{pid, rand.Float64()})
	assert.NilError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "http://example.com/asset/"+pid.String()+"/status", bytes.NewBuffer(reqBody))

	router := makeRouter()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code, "post returned 201")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")

}

func TestAssetControlGet(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/asset/"+pid.String()+"/control", nil)

	router := makeRouter()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "get returned 200")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")

	ctrl := AssetControl{}
	err = json.Unmarshal(w.Body.Bytes(), &ctrl)

	assert.NilError(t, err)
	assert.Equal(t, pid, ctrl.PID)
}
