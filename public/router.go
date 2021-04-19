package public

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Router provides HHTP endpoints to access the schedule
func (p *Publicer) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/channels", p.GetChannelsHandler).Methods("GET")
	r.HandleFunc("/channel/{name}", p.GetChannelHandler).Methods("GET")
	return r
}

// GetChannelsHander handles requesting all channels
func (p *Publicer) GetChannelsHandler(w http.ResponseWriter, r *http.Request) {
	chs, err := p.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&chs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetChannelHandler handles requesting a single channel's information
func (p *Publicer) GetChannelHandler(w http.ResponseWriter, r *http.Request) {
	channelID := mux.Vars(r)["name"]
	if len(channelID) == 0 {
		http.Error(w, "invalid channel name", http.StatusBadRequest)
	}
	ch, err := p.GetChannel(r.Context(), channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&ch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("public"))
}
