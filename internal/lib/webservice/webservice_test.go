package webservice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"gotest.tools/assert"
)

func TestAssetStatusGet(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	app := App{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/asset/"+pid.String()+"/status", nil)

	router := app.Router()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code, "get returned 200")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")
}

type testAssetStatus struct {
	Test1 string  `json:"Test1"`
	Test2 float64 `json:"Test2"`
	Test3 bool    `json:"Test3"`
}

func TestAssetStatusPost(t *testing.T) {
	pid, err := uuid.NewUUID()
	assert.NilError(t, err)

	app := App{}

	status := testAssetStatus{
		Test1: "hi",
		Test2: 2.1,
		Test3: true,
	}

	reqBody, err := json.Marshal(status)

	var statusReturn testAssetStatus
	err = json.Unmarshal(reqBody, &statusReturn)
	assert.NilError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "http://example.com/asset/"+pid.String()+"/status", bytes.NewBuffer(reqBody))

	router := app.Router()
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code, "post returned 201")
	assert.Equal(t, "application/json; charset=UTF-8", w.HeaderMap.Get("Content-Type"), "got expected Content-Type in response")

}

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
