package web

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/msg"
	"gotest.tools/assert"
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

func TestPublish(t *testing.T) {
	handler, _ := New("./web_handler_config.json")

	pid, _ := uuid.NewUUID()
	ch := make(chan msg.Msg)
	handler.Publish(ch)

	data := []byte("hi")
	ch <- msg.New(pid, data)

	assert.Equal(t, handler.msgs[0], data)
}
