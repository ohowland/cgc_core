package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ohowland/cgc/internal/pkg/asset"
)

var clients = make(map[*websocket.Conn]bool)
var command = make(chan CommandMessage)
var upgrader = websocket.Upgrader{}

type UpdateMessage struct {
	Asset string `json:"asset"`
	KW    string `json:"kw"`
	KVAR  string `json:"kvar"`
}

type CommandMessage struct {
	Asset string `json:"asset"`
	Run   bool   `json:"run"`
	KW    string `json:"kw"`
	KVAR  string `json:"kvar"`
}

func StartServer(assets map[uuid.UUID]asset.Asset) {

	fs := http.FileServer(http.Dir("./internal/pkg/web/static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go hmiUpdater(assets)
	go commandDistribution(assets)

	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg CommandMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		command <- msg
	}
}

func hmiUpdater(assets map[uuid.UUID]asset.Asset) {
	// send status messages to client
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		for _, a := range assets {
			power := a.(asset.PowerReader)
			msg := UpdateMessage{
				Asset: a.Name(),
				KW:    fmt.Sprintf("%f", power.KW()),
				KVAR:  fmt.Sprintf("%f", power.KVAR()),
			}
			for client := range clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

func commandDistribution(assets map[uuid.UUID]asset.Asset) {
	assetControllers := machineControlMap(assets)
	for msg := range command {
		c, ok := assetControllers[msg.Asset]
		if !ok {
			log.Printf("unknown asset name: %v\n", msg.Asset)
		} else {
			c.RunCmd(msg.Run)
			kw, err := strconv.ParseFloat(msg.KW, 64)
			if err == nil {
				c.KWCmd(kw)
			}
			kvar, err := strconv.ParseFloat(msg.KVAR, 64)
			if err == nil {
				c.KVARCmd(kvar)
			}
		}
	}
}

func machineControlMap(assets map[uuid.UUID]asset.Asset) map[string]asset.MachineControl {
	controlMap := make(map[string]asset.MachineControl)
	for _, asset := range assets {
		controlMap[asset.Name()] = asset.DispatchControlHandle()
	}
	return controlMap
}
