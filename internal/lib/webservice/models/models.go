package models

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "zx125!G33"
	dbname   = "cgc_test"
)

type AssetStatus struct {
	KW   float64
	KVAR float64
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
		pid UUID PRIMARY KEY,
		kw FLOAT,
		kvar FLOAT
	);`
	_, err = db.Exec(statusTable)
	if err != nil {
		panic(err)
	}

	controlTable := `
	CREATE TABLE asset_control (
		pid UUID PRIMARY KEY,
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
