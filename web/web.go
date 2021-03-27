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

	web.mux.HandleFunc("/", web.indexHandle)
	return web
}

func (web *Web) Router() *mux.Router {
	return web.mux
}

func (web *Web) indexHandle(w http.ResponseWriter, r *http.Request) {
	chs, err := web.mcr.GetChannels()
	tempChan := []templates.Channel{}
	for _, ch := range chs {
		tempChan = append(tempChan, templates.Channel{
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
		Channels: tempChan,
	}

	err = web.t.Dashboard(w, params)
	if err != nil {
		err = fmt.Errorf("failed to render dashboard: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
