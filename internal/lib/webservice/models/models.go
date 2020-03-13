package models

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "cgc"
	password = "x9xwa25"
	dbname   = "cgc_test"
)

type AssetStatus struct {
	PID  uuid.UUID `json:"PID"`
	Name string    `json:"Name"`
	KW   float64   `json:"KW"`
	KVAR float64   `json:"KVAR"`
}

// NewDB returns a database reference
func NewDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)

	return db, err
}

// createTables initializes the database schema
func createTables() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	statusTable := `
	CREATE TABLE asset_status (
	    name VARCHAR(257) PRIMARY KEY,
		pid UUID,
		kw FLOAT,
		kvar FLOAT
	);`
	_, err = db.Exec(statusTable)
	if err != nil {
		panic(err)
	}

	controlTable := `
	CREATE TABLE asset_control (
	    name VARCHAR(256) PRIMARY KEY,
		pid UUID,
		run_request BOOL,
		kw FLOAT,
		kvar FLOAT,
		gridform BOOL
	);`
	_, err = db.Exec(controlTable)
	if err != nil {
		panic(err)
	}
}
