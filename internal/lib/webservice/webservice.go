package webservice

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ohowland/cgc/internal/lib/webservice/models"
)

type Config struct {
	URL  string
	Port string
}

type App struct {
	DB     *sql.DB
	Config Config
}

func (app *App) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", app.BaseHandler)
	r.HandleFunc("/assets/status", app.AllStatusHandler).Methods("GET")
	r.HandleFunc("/assets/{pid}/status", app.StatusHandler).Methods("GET", "POST")
	//r.HandleFunc("/asset/{pid}/control", app.ControlHandler).Methods("GET")
	return r
}

func wrapHandler(handler func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	h := func(w http.ResponseWriter, r *http.Request) {
		// wrap goes here
		handler(w, r)
	}
	return h
}

func (app *App) BaseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (app *App) AllStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	rows, err := app.DB.Query(`SELECT (name, pid, kw, kvar) FROM asset_status`)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	defer rows.Close()

	var statusArray []models.AssetStatus
	for rows.Next() {
		status := models.AssetStatus{}
		err = rows.Scan(&status.Name, &status.PID, &status.KW, &status.KVAR)
		if err == nil {
			statusArray = append(statusArray, status)
		}
	}

	body, err := json.Marshal(statusArray)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)

}

func (app *App) StatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	switch r.Method {
	case "GET":
		pid, err := uuid.Parse(vars["pid"])
		if err != nil {
			log.Println("malformed UUID:", err)
		}

		rows, err := app.DB.Query(`SELECT (name, kw, kvar) FROM asset_status WHERE pid = $1`, pid)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			defer rows.Close()
			status := models.AssetStatus{}
			for rows.Next() {
				err = rows.Scan(&status.Name, &status.KW, &status.KVAR)
			}

			body, err := json.Marshal(status)
			if err != nil {
				log.Println("malformed JSON:", err)
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(body)
		}

	case "POST":
		w.WriteHeader(http.StatusCreated)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		req := make(map[string]interface{})
		err = json.Unmarshal(body, &req)

		sqlStatement := `INSERT INTO asset_status (name, pid, kw, kvar) 
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (name) DO UPDATE SET
				 	pid = $2,
				     	kw = $3,
				     	kvar = $4;`

		_, err = app.DB.Exec(sqlStatement, req["Name"], req["PID"], req["KW"], req["KVAR"])

		log.Println("POSTED to Status:", req)
		log.Println(err)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
