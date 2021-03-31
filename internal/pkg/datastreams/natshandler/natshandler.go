package natshandler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"

	nats "github.com/nats-io/nats.go"
)

type Handler struct {
	mux    *sync.Mutex
	inbox  <-chan msg.Msg
	pid    uuid.UUID
	config config
	stop   chan bool
}

type config struct {
	Server string `json:"Server"`
}

func (h Handler) PID() uuid.UUID {
	return h.pid
}

func redirectMsg(chIn <-chan msg.Msg, chOut chan<- msg.Msg) {
	for m := range chIn {
		chOut <- m
	}
}

func New(configPath string, system msg.Publisher) (Handler, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Handler{}, err
	}
	cfg := config{}
	if err := json.Unmarshal(jsonConfig, &cfg); err != nil {
		return Handler{}, err
	}

	pid, _ := uuid.NewUUID()

	inbox := make(chan msg.Msg, 50)

	chStatus, err := system.Subscribe(pid, msg.Status)
	if err != nil {
		panic(err)
	}
	go redirectMsg(chStatus, inbox)

	chConfig, err := system.Subscribe(pid, msg.Config)
	if err != nil {
		panic(err)
	}
	go redirectMsg(chConfig, inbox)

	if err := json.Unmarshal(jsonConfig, &cfg); err != nil {
		return Handler{}, err
	}

	stop := make(chan bool)

	return Handler{
		mux:    &sync.Mutex{},
		inbox:  inbox,
		pid:    pid,
		config: cfg,
		stop:   stop,
	}, nil
}

func (h *Handler) Stop() {
	h.stop <- true
}

func (h Handler) Process() {
	log.Println("[NATS client] Process Started")
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

loop:
	for {
		select {
		case m := <-h.inbox:
			switch m.Topic() {
			case msg.Status:
				data, err := json.Marshal(m.Payload())
				if err != nil {
					continue
				}
				if err = nc.Publish(m.PID().String(), data); err != nil {
					log.Printf("unable to publish to nats server: %v", err)
				}

			case msg.Config:
			}

		case <-h.stop:
			nc.Close()
			break loop
		}
	}
	log.Println("[NATS client] Process Shutdown")
}
