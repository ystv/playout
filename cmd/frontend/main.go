//Package main this version will only load the frontend
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	// PostgreSQL driver
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/ystv/playout/channel"
	"github.com/ystv/playout/playout"
	"github.com/ystv/playout/programming"
	"github.com/ystv/playout/public"
	"github.com/ystv/playout/web"
)

func main() {
	db, err := newDatabase()
	if err != nil {
		log.Fatalf("failed to start db: %+v", err)
	}
	mcr, err := channel.NewMCR(db)
	if err != nil {
		log.Fatalf("failed to create mcr: %+v", err)
	}
	prog := programming.New(db)
	po := playout.New(prog, db)
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")
	mount(r, "/playout", web.New(mcr).Router())
	mount(r, "/public", public.New(mcr, prog, po).Router())

	log.Fatal(http.ListenAndServe("0.0.0.0:7070", r))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("playout (v0.0.3) frontend only (/playout)"))
	w.WriteHeader(http.StatusOK)
}

// mount another mux router ontop of another
func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
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
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return db, nil
}
