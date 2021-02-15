package playout

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Router provides HHTP endpoints to access the schedule
func (po *Playouter) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	// TODO: Add scheduler endpoints
	r.HandleFunc("/playouts", po.newPlayout).Methods("POST")
	r.HandleFunc("/playouts", po.updatePlayout).Methods("PUT")
	return r
}

func (po *Playouter) newPlayout(w http.ResponseWriter, r *http.Request) {
	playout := NewPlayout{}
	err := json.NewDecoder(r.Body).Decode(&playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	playoutID, err := po.New(r.Context(), playout)
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

func (po *Playouter) updatePlayout(w http.ResponseWriter, r *http.Request) {
	playout := Playout{}
	err := json.NewDecoder(r.Body).Decode(&playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = po.Update(r.Context(), playout)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("scheduler"))
}
