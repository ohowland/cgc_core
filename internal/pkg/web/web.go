package web

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
)

type handler struct {
	mux    *sync.Mutex
	pid    uuid.UUID
	msgs   []json.RawMessage
	config config
}

type config struct {
	URL string
}

// thinking this gets linked into the asset broadcast
// so handler would call subscribe and link the channel to
// publish method, which aggregates

func New(configPath string) (handler, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return handler{}, err
	}
	cfg := config{}
	if err := json.Unmarshal(jsonConfig, &cfg); err != nil {
		return handler{}, err
	}

	pid, _ := uuid.NewUUID()
	msgs := make([]json.RawMessage, 0)

	return handler{
		mux:    &sync.Mutex{},
		pid:    pid,
		msgs:   msgs,
		config: cfg}, err
}

func (h *handler) Publish(ch <-chan msg.Msg) {
	go func() {
		for {
			data := <-ch
			switch data.Topic() {
			case msg.JSON:
				bytes, ok := data.Payload().(json.RawMessage)
				if !ok {
					continue
				}
				h.enqueue(bytes)

			default:
			}
		}
	}()
}

func (h *handler) enqueue(bytes json.RawMessage) {
	h.mux.Lock()
	defer h.mux.Unlock()
	h.msgs = append(h.msgs, bytes)
}

func (h *handler) transport() {
	go func() {
		for {
			if len(h.msgs) > 0 {
				send := h.dequeue()
				for 
				h.PostAssetStatus(send)
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

func (h *handler) dequeue() []json.RawMessage {
	h.mux.Lock()
	defer h.mux.Unlock()

	var send []json.RawMessage
	copy(send[:], h.msgs)
	h.msgs = nil
	return send
}

func (h handler) PostAssetStatus(bytes ) {
	name := "hmm"
	targetURL := h.config.URL + "/assets/" + name + "/status"
	//log.Println("TARGET:", targetURL)
	//log.Println("JSON:", b)
	_, err := http.Post(targetURL, "Content-Type: application/json", bytes.NewBuffer(bytes))
	if err != nil {
		log.Println("[Webservice Handler]", err)
	}
}

/*
func (h Handler) updateHandler() {s
	for pid, status := range w.GetStatus() {

		var b []byte
		var err error
		switch s := status.(type) {
		case ess.Status:
			b, err = json.Marshal(s)
		case grid.Status:
			b, err = json.Marshal(s)
		case feeder.Status:
			b, err = json.Marshal(s)
		case pv.Status:
			b, err = json.Marshal(s)
		default:
			b = json.RawMessage("{}")
			err = errors.New("manualDispatch.updateHandler: Could not cast status")
		}

		if err != nil {
			log.Println(err)
		}

		targetURL := h.config.URL + "/assets/" + pid.String() + "/status"
		log.Println("TARGET:", targetURL)
		log.Println("JSON:", b)
		_, err = http.Post(targetURL, "Content-Type: application/json", bytes.NewBuffer(b))
	}
}
*/
