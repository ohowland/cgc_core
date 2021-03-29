package mongodb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	mux    *sync.Mutex
	inbox  <-chan msg.Msg
	pid    uuid.UUID
	config config
	stop   chan bool
}

type config struct {
	URI      string `json:"URI"`
	Database string `json:"Database"`
	Port     string `json:"Port"`
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
	go redirectMsg(chStatus, inbox)

	chConfig, err := system.Subscribe(pid, msg.Config)
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

func msgToBSON(m msg.Msg) bson.D {
	//TODO: PID should be written as a binary of subtype 0x04 (UUID standard).
	// currently written as a string.
	return bson.D{
		{"$set", bson.M{
			"pid":  m.PID().String(),
			"data": m.Payload(),
		}},
	}
}

func (h *Handler) StopProcess() {
	h.stop <- true
}

func (h Handler) Process() {
	//TODO: Handle reconnection to the MongoDB resource
	client, err := mongo.NewClient(options.Client().ApplyURI(h.config.URI + ":" + h.config.Port))
	if err != nil {
		log.Println(err)
	}

	//ctx, _ := context.WithTimeout(context.TODO(), 20*time.Second)
	ctx := context.TODO()
	err = client.Connect(ctx)
	if err != nil {
		log.Println(err)
	}
	defer client.Disconnect(ctx)

	client.Database(h.config.Database).Collection("assetStatus").Drop(ctx)
	client.Database(h.config.Database).Collection("assetConfig").Drop(ctx)
loop:
	for {
		select {
		case m := <-h.inbox:
			switch m.Topic() {
			case msg.Status:
				opts := options.Update().SetUpsert(true)
				_, err = client.Database(h.config.Database).Collection("assetStatus").UpdateOne(
					ctx,
					bson.M{"pid": m.PID()},
					msgToBSON(m),
					opts,
				)

				if err != nil {
					log.Fatal(err)
				}

			case msg.Config:
				log.Println("[Mongo] Config:", m)
				opts := options.Update().SetUpsert(true)
				_, err = client.Database(h.config.Database).Collection("assetConfig").UpdateOne(
					ctx,
					bson.M{"pid": m.PID()},
					msgToBSON(m),
					opts,
				)

				if err != nil {
					log.Fatal(err)
				}
			}
		case <-h.stop:
			break loop
		}
	}
	log.Println("[Mongo] Process Shutdown")
}
