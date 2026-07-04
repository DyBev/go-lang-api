package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

type App struct {
	templates *template.Template
}

func main() {
	tmpl := template.Must(template.ParseGlob("templates/**/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))

	app := &App{templates: tmpl}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.index)
	mux.HandleFunc("/partials/time", app.timePartial)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logging(mux)))
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := map[string]any{
		"Title": "Go + HTMX",
		"Now":   time.Now().Format(time.RFC1123),
	}
	if err := a.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (a *App) timePartial(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Now": time.Now().Format(time.RFC1123),
	}
	if err := a.templates.ExecuteTemplate(w, "time.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
