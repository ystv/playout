package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/reiver/go-telnet"
)

var lastSelection string
var tmpl *template.Template

func main() {
	var err error
	tmpl = template.New("index")
	tmpl, err = template.ParseFiles("static/index.html")
	if err != nil {
		log.Fatalf("failed to load templates: %+v", err)
	}

	conn, err := telnet.DialTo("127.0.0.1:1234")
	if err != nil {
		log.Fatalf("failed to connect to liquidsoap: %+v", err)
	}

	r := mux.NewRouter()
	r.Handle("/", BasicAuth(handleIndex, "hackme", "hackme", "YSTV Playout")).Methods("GET")

	r.HandleFunc("/select/{feed}", func(w http.ResponseWriter, r *http.Request) {
		req := strings.Split(mux.Vars(r)["feed"], ".")
		feed := req[len(req)-1]
		conn.Write([]byte("playout.select." + feed))
		conn.Write([]byte("\n"))
		log.Printf("switched to feed: %s", feed)
		lastSelection = feed
		http.Redirect(w, r, "/", http.StatusFound)
	}).Methods("POST")

	log.Fatal(http.ListenAndServe("0.0.0.0:7070", r))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := tmpl.Execute(w, lastSelection)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func BasicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}
