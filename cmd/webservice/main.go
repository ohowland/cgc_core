package main

import (
	"log"
	"net/http"

	"github.com/ohowland/cgc/internal/lib/webservice"
	"github.com/ohowland/cgc/internal/lib/webservice/models"
)

func main() {
	db, err := models.NewDB()
	if err != nil {
		panic(err)
	}

	app := webservice.App{
		DB: db,
	}

	r := app.Router()
	http.Handle("/", r)

	port := ":8080"
	log.Println("Starting Server on Port", port)
	http.ListenAndServe(port, r)
}
