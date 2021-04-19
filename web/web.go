package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ystv/playout/channel"
	"github.com/ystv/playout/web/templates"
)

type Web struct {
	mux *mux.Router
	mcr *channel.MCR
	t   *templates.Templater
}

func New(mcr *channel.MCR) *Web {
	web := &Web{mux: mux.NewRouter(), mcr: mcr, t: templates.New()}

	web.mux.HandleFunc("/", web.indexPage).Methods("GET")
	web.mux.HandleFunc("/ch/{channel}", web.channelPage).Methods("GET")
	web.mux.HandleFunc("/channel/new", web.newChannelPage).Methods("GET")
	web.mux.HandleFunc("/channel/new", web.newChannel).Methods("POST")
	web.mux.HandleFunc("/settings", web.settingsPage).Methods("GET")

	return web
}

func (web *Web) Router() *mux.Router {
	return web.mux
}

func (web *Web) indexPage(w http.ResponseWriter, r *http.Request) {
	chs, err := web.mcr.GetChannels()
	tempChans := []templates.Channel{}
	for _, ch := range chs {
		tempChans = append(tempChans, templates.Channel{
			ShortName:   ch.ShortName,
			ChannelType: ch.ChannelType,
			IngestURL:   ch.IngestURL,
			IngestType:  ch.IngestType,
			SlateURL:    ch.SlateURL,
			Archive:     ch.Archive,
			Status:      ch.Status,
			Name:        ch.Name,
			Description: ch.Description,
			Thumbnail:   ch.Thumbnail,
			CreatedAt:   ch.CreatedAt,
		})
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := templates.DashboardParams{
		Base: templates.BaseParams{
			UserName:   "rhys",
			SystemTime: time.Now(),
		},
		Channels: tempChans,
	}

	err = web.t.Dashboard(w, params)
	if err != nil {
		err = fmt.Errorf("failed to render dashboard: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (web *Web) channelPage(w http.ResponseWriter, r *http.Request) {
	ch, err := web.mcr.GetChannel(r.Context(), mux.Vars(r)["channel"])
	if err != nil {
		err = fmt.Errorf("failed to get channel: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	params := templates.ChannelParams{
		Base: templates.BaseParams{
			UserName:   "rhys",
			SystemTime: time.Now(),
		},
		Ch: templates.Channel{
			ShortName:   ch.ShortName,
			ChannelType: ch.ChannelType,
			IngestURL:   ch.IngestURL,
			IngestType:  ch.IngestType,
			SlateURL:    ch.SlateURL,
			Archive:     ch.Archive,
			Status:      ch.Status,
			Name:        ch.Name,
			Description: ch.Description,
			Thumbnail:   ch.Thumbnail,
			CreatedAt:   ch.CreatedAt,
		},
	}
	err = web.t.Channel(w, params)
	if err != nil {
		err = fmt.Errorf("failed to render channel page: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (web *Web) newChannelPage(w http.ResponseWriter, r *http.Request) {
	params := templates.PlainParams{
		Base: templates.BaseParams{
			UserName:   "rhys",
			SystemTime: time.Now(),
		},
	}
	err := web.t.NewChannel(w, params)
	if err != nil {
		err = fmt.Errorf("failed to render new channel page: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (web *Web) newChannel(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		err = fmt.Errorf("failed to parse form: %w", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	isArchived := false
	if r.PostFormValue("archive") != "" {
		isArchived = true
	}
	isDVR := false
	if r.PostFormValue("vcr") != "" {
		isDVR = true
	}
	newCh := channel.NewChannelStruct{
		Name:        r.PostFormValue("name"),
		ShortName:   r.PostFormValue("short-name"),
		Description: r.PostFormValue("description"),
		ChannelType: r.PostFormValue("type"),
		IngestType:  r.PostFormValue("ingest-type"),
		SlateURL:    r.PostFormValue("slate-url"),
		Archive:     isArchived,
		DVR:         isDVR,
	}
	ch, err := web.mcr.NewChannel(r.Context(), newCh)
	if err != nil {
		err = fmt.Errorf("failed to create channel: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/playout/ch/"+ch.ShortName, http.StatusFound)
}

func (web *Web) settingsPage(w http.ResponseWriter, r *http.Request) {
	params := templates.PlainParams{
		Base: templates.BaseParams{
			UserName:   "rhys",
			SystemTime: time.Now(),
		},
	}
	err := web.t.Settings(w, params)
	if err != nil {
		err = fmt.Errorf("failed to render settings page: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
