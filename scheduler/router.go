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
	r.HandleFunc("/block", s.newBlock).Methods("POST")
	r.HandleFunc("/block", s.updateBlock).Methods("PUT")
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
	blockID, err := s.NewBlock(r.Context(), block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := struct {
		blockID int `json:"blockID"`
	}{
		blockID: blockID,
	}

	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *Scheduler) updateBlock(w http.ResponseWriter, r *http.Request) {
	block := Block{}
	err := json.NewDecoder(r.Body).Decode(&block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.UpdateBlock(r.Context(), block)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
