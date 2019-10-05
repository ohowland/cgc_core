package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var upgrader = websocket.Upgrader{}

type Message struct {
	Asset string `json:"asset"`
	KW    string `json:"kw"`
	KVAR  string `json:"kvar"`
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", handleConnections)

	go pingMessages()

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
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// send control messages back to controller
	}
}

func pingMessages() {
	// send status messages to client
	ticker := time.NewTicker(200 * time.Millisecond)
	assets := []string{"ess", "grid", "feeder"}
	for range ticker.C {
		kw := strconv.Itoa(rand.Intn(100))
		kvar := strconv.Itoa(rand.Intn(100))
		msg := Message{
			Asset: assets[rand.Intn(3)],
			KW:    kw,
			KVAR:  kvar,
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
