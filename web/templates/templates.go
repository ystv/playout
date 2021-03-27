package templates

import (
	"embed"
	"html/template"
	"io"
	"time"
)

//go:embed *.tmpl
var tpls embed.FS

type (
	Templater struct {
		dashboard *template.Template
		funcs     template.FuncMap
	}
	Channel struct {
		ShortName   string // URL name
		ChannelType string // Event / linear
		IngestURL   string
		IngestType  string
		SlateURL    string
		Archive     bool
		Status      string
		Name        string // Display name
		Description string
		Thumbnail   string
		CreatedAt   time.Time
	}
	BaseParams struct {
		UserName   string
		SystemTime time.Time
	}
	DashboardParams struct {
		Base     BaseParams
		Channels []Channel
	}
)

var funcs = template.FuncMap{
	"cleantime": cleanTime,
}

func cleanTime(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func parse(file string) *template.Template {
	return template.Must(
		template.New("base.tmpl").Funcs(funcs).ParseFS(tpls, "base.tmpl", file))
}

func New() *Templater {
	return &Templater{
		dashboard: parse("dashboard.tmpl"),
	}
}

func (t *Templater) Dashboard(w io.Writer, p DashboardParams) error {
	return t.dashboard.Execute(w, p)
}
