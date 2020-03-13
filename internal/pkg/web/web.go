package web

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type handler struct {
	config config
}

type config struct {
	URL string
}

func New(configPath string) (handler, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return handler{}, err
	}
	cfg := config{}
	if err := json.Unmarshal(jsonConfig, &cfg); err != nil {
		return handler{}, err
	}
	return handler{config: cfg}, err
}

func (h handler) PostAssetStatus(name string, jsonData []byte) {
	targetURL := h.config.URL + "/assets/" + name + "/status"
	//log.Println("TARGET:", targetURL)
	//log.Println("JSON:", b)
	_, err := http.Post(targetURL, "Content-Type: application/json", bytes.NewBuffer(jsonData))
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
