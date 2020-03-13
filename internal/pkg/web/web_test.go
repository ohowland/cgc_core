package web

import (
	"testing"
)

func TestNewHandler(t *testing.T) {
	h, err := New("./web_handler_config.json")

	if err != nil {
		t.Error(err)
	}

	expectedURL := "http://192.168.0.5"
	if h.config.URL != expectedURL {
		t.Errorf("New(): FAILED. Expected URL %v, but got URL %v", expectedURL, h.config.URL)
	} else {
		t.Logf("New(): PASSED. Expected URL %v, and got URL %v", expectedURL, h.config.URL)
	}
}

func TestSubscribe(t *testing.T)
