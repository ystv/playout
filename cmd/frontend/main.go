//Package main this version will only load the frontend
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ystv/playout/web"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")
	mount(r, "/playout", web.New())

	log.Fatal(http.ListenAndServe("0.0.0.0:7070", r))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("playout (v0.0.2) frontend only (/playout)"))
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
