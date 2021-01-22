package scheduler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Router provides HHTP endpoints to access the schedule
func (s *Scheduler) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/new_block", s.newBlock).Methods("POST")
	return r
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("scheduler"))
}

func (s *Scheduler) newBlock(w http.ResponseWriter, r *http.Request) {
	block := NewBlock{}
	err := json.NewDecoder(r.Body).Decode(&block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.NewBlock(r.Context(), block)
}
