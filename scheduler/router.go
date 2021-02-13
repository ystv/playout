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
	r.HandleFunc("/playout", s.newPlayout).Methods("POST")
	r.HandleFunc("/playout", s.updatePlayout).Methods("PUT")
	return r
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("scheduler"))
}

func (s *Scheduler) newPlayout(w http.ResponseWriter, r *http.Request) {
	playout := NewPlayout{}
	err := json.NewDecoder(r.Body).Decode(&playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	playoutID, err := s.NewPlayout(r.Context(), playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := struct {
		PlayoutID int `json:"playoutID"`
	}{
		PlayoutID: playoutID,
	}

	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
}

func (s *Scheduler) updatePlayout(w http.ResponseWriter, r *http.Request) {
	playout := Playout{}
	err := json.NewDecoder(r.Body).Decode(&playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.UpdatePlayout(r.Context(), playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
