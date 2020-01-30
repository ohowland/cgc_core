package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ohowland/cgc/internal/pkg/webservice"
)

func makeRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", webservice.BaseHandler)
	r.HandleFunc("/asset/{pid}/status", webservice.StatusHandler).Methods("GET", "POST")
	r.HandleFunc("/asset/{pid}/control", webservice.ControlHandler).Methods("GET")
	return r
}

func main() {
	r := makeRouter()
	http.Handle("/", r)

	port := ":8080"
	log.Println("Starting Server on Port", port)
	http.ListenAndServe(port, r)
}
