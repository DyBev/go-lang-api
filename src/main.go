package main

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

var openConns int64

type App struct {
	templates *template.Template
	taskID int
	hub *Hub
}

func main() {
	tmpl := template.Must(template.ParseGlob("templates/**/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("templates/*.html"))

	app := &App{
		templates: tmpl,
		taskID: 0,
		hub: NewHub(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.index)
	mux.HandleFunc("/events/hub", app.eventsSSE)
	mux.HandleFunc("PUT /api/register-task", app.registerTask)
	mux.HandleFunc("PATCH /api/complete-task/{taskID}", app.completeTask)
	mux.HandleFunc("GET /api/heartbeat", app.clientHeartbeat)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	go logRuntimeStats(10 * time.Second)
	go app.timeSSE()
	go cleanupLoop()

	srv := &http.Server{
		Addr: ":8080",
		Handler: logging(clientIDMiddleware(mux)),

		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				atomic.AddInt64(&openConns, 1)
			case http.StateHijacked, http.StateClosed:
				atomic.AddInt64(&openConns, -1)
			}
		},
	}

	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
