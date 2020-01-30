package webservice

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AssetStatus struct {
	PID    uuid.UUID `json:"PID"`
	Status float64   `json:"Data"`
}

type AssetControl struct {
	PID     uuid.UUID `json:"PID"`
	Control float64   `json:"Data"`
}

func wrapHandler(handler func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	h := func(w http.ResponseWriter, r *http.Request) {
		// wrap goes here
		handler(w, r)
	}
	return h
}

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	switch r.Method {
	case "GET":
		pid, err := uuid.Parse(vars["pid"])
		if err != nil {
			log.Println("malformed UUID:", err)
		}
		resp := AssetStatus{
			PID:    pid,
			Status: 123.456,
		}

		body, err := json.Marshal(resp)
		if err != nil {
			log.Println("malformed JSON:", err)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(body)

	case "POST":
		w.WriteHeader(http.StatusCreated)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		status := AssetStatus{}
		if err = json.Unmarshal(body, &status); err != nil {
			log.Println("malformed JSON:", err)
		}

		log.Println("POSTED to Status:", status)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func ControlHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	switch r.Method {
	case "GET":
		resp := struct {
			PID  string  `json:"PID"`
			Data float64 `json:"Data"`
		}{
			PID:  vars["pid"],
			Data: 987.654,
		}

		body, err := json.Marshal(resp)
		if err != nil {
			log.Println("malformed JSON:", err)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(body)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
