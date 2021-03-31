package sqldb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ohowland/cgc_core/internal/pkg/msg"

	_ "github.com/go-sql-driver/mysql"
)

type Handler struct {
	mux    *sync.Mutex
	inbox  <-chan msg.Msg
	pid    uuid.UUID
	config config
	stop   chan bool
}

type config struct {
	Server   string `json:"Server"`
	Port     int    `json:"Port"`
	Username string `json:"Username"`
	Password string `json:"Password"`
	Database string `json:"Database"`
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

func (h Handler) DB() (*sql.DB, error) {
	uri := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", h.config.Username, h.config.Password, h.config.Server, h.config.Port, h.config.Database)
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (h Handler) Process() {
	db, err := h.DB()
	defer db.Close()
	if err != nil {
		panic(err) // #TODO Handle failed connection
	}

	err = initDBTables(db)
	if err != nil {
		panic(err) // #TODO Handle failed query
	}

loop:
	for {
		select {
		case m := <-h.inbox:
			switch m.Topic() {
			case msg.Status:
				sqlStatement := `INSERT INTO realtime VALUES (uuid $1, status $2)`

				ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
				_, err := db.ExecContext(ctx, sqlStatement, m.PID().String, m.Payload())
				if err != nil {
					log.Printf("error %s update db", err)
				}

			case msg.Config:
			}

		case <-h.stop:
			break loop
		}
	}
	log.Println("[Mongo] Process Shutdown")
}

func initDBTables(db *sql.DB) error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS realtime(uuid VARCHAR(32) primary key, status BLOB, config BLOB)`
	_, err := db.Exec(sqlStatement)
	return err
}

func updateRow(db *sql.DB, m msg.Msg) {
	sqlStatement := `INSERT INTO realtime (uuid, status) VALUES ($1, $2) ON DUPLICATE KEY UPDATE`
	_, err := db.Exec(sqlStatement, m.PID().String())
}
