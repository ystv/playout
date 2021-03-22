package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ystv/playout/public"
	"github.com/ystv/playout/web/templates"
)

func New() *mux.Router {
	m := mux.NewRouter()
	t := templates.New()

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		params := templates.DashboardParams{
			Base: templates.BaseParams{
				UserName:   "rhys",
				SystemTime: time.Now(),
			},
			Channels: []templates.Channel{
				{
					public.Channel{
						ShortName:    "One",
						Name:         "York Student Television One",
						Description:  "Something",
						Destinations: []string{"somewhere", "rah"},
					},
				},
			},
		}

		err := t.Dashboard(w, params)
		if err != nil {
			err = fmt.Errorf("failed to render dashboard: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	return m
}
