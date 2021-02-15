package scheduler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router provides HHTP endpoints to access the schedule
func (s *Scheduler) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	// TODO: Add scheduler endpoints
	return r
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("scheduler"))
}
