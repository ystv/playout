package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	// PostgreSQL driver
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/ystv/playout/scheduler"
)

func main() {
	log.Println("playout (v0.0.2) by Rhys Milling")
	db, err := newDatabase()
	if err != nil {
		log.Fatalf("failed to start DB: %+v", err)
	}
	s, err := scheduler.New(db)
	if err != nil {
		log.Fatalf("failed to start scheduler: %+v", err)
	}

	err = s.MainLoop(context.Background())
	if err != nil {
		log.Fatalf("scheduler failed: %+v", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")

	log.Fatal(http.ListenAndServe("0.0.0.0:7070", r))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("playout (v0.0.2)"))
	w.WriteHeader(http.StatusOK)
}

// newDatabase creates a new database connection
func newDatabase() (*sqlx.DB, error) {
	username := os.Getenv("PLAYOUT_DB_USER")
	password := os.Getenv("PLAYOUT_DB_PASS")
	dbName := os.Getenv("PLAYOUT_DB_NAME")
	dbHost := os.Getenv("PLAYOUT_DB_HOST")
	dbPort := os.Getenv("PLAYOUT_DB_PORT")

	dbURI := fmt.Sprintf("dbname=%s host=%s user=%s password=%s port=%s sslmode=disable", dbName, dbHost, username, password, dbPort) // Build connection string

	db, err := sqlx.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	return db, nil
}
