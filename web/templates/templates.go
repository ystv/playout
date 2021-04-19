package templates

import (
	"embed"
	"html/template"
	"io"
	"time"
)

//go:embed *.tmpl
//go:embed channel/*.tmpl
var tpls embed.FS

type (
	Templater struct {
		dashboard  *template.Template
		channel    *template.Template
		newChannel *template.Template
		settings   *template.Template
		funcs      template.FuncMap
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
	PlainParams struct {
		Base BaseParams
	}
	BaseParams struct {
		UserName   string
		SystemTime time.Time
	}
	DashboardParams struct {
		Base     BaseParams
		Channels []Channel
	}
	ChannelParams struct {
		Base BaseParams
		Ch   Channel
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
		dashboard:  parse("dashboard.tmpl"),
		channel:    parse("channel/channel.tmpl"),
		newChannel: parse("channel/new-channel.tmpl"),
		settings:   parse("settings-main.tmpl"),
	}
}

func (t *Templater) Dashboard(w io.Writer, p DashboardParams) error {
	return t.dashboard.Execute(w, p)
}

func (t *Templater) Channel(w io.Writer, p ChannelParams) error {
	return t.channel.Execute(w, p)
}

func (t *Templater) NewChannel(w io.Writer, p PlainParams) error {
	return t.newChannel.Execute(w, p)
}

func (t *Templater) Settings(w io.Writer, p PlainParams) error {
	return t.settings.Execute(w, p)
}
