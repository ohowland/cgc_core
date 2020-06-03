package web

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
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

	data := json.RawMessage("{name: owen}")
	ch <- msg.New(pid, msg.JSON, data)

	var err error
	for i := range data {
		if data[i] != handler.msgs[0][i] {
			t.Errorf("Publish(): FAILED. %v != %v", data[i], handler.msgs[0][i])
			err = errors.New("not equal")
		}
	}
	if err == nil {
		t.Logf("Publish(): PASSED. %v == %v ", data, handler.msgs[0])
	}
}

func TestPublishMulti(t *testing.T) {
	handler, _ := New("./web_handler_config.json")

	pid, _ := uuid.NewUUID()
	ch := make(chan msg.Msg)
	handler.Publish(ch)

	data1 := json.RawMessage("{name: owen, job: keyboard typing}")
	data2 := json.RawMessage("{name: howland}")
	ch <- msg.New(pid, msg.JSON, data1)
	ch <- msg.New(pid, msg.JSON, data2)

	var err error
	for i := range data1 {
		if data1[i] != handler.msgs[0][i] {
			t.Errorf("Publish(): FAILED. %v != %v", data1[i], handler.msgs[0][i])
			err = errors.New("not equal")
		}
	}
	if err == nil {
		t.Logf("Publish(): PASSED. %v == %v ", data1, handler.msgs[0])
	}

	for i := range data2 {
		if data2[i] != handler.msgs[1][i] {
			t.Errorf("Publish(): FAILED. %v != %v", data2[i], handler.msgs[1][i])
			err = errors.New("not equal")
		}
	}
	if err == nil {
		t.Logf("Publish(): PASSED. %v == %v ", data2, handler.msgs[1])
	}
}
