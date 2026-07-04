package main

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"runtime"
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

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	go logRuntimeStats(10 * time.Second)
	go app.timeSSE()

	srv := &http.Server{
		Addr: ":8080",
		Handler: logging(mux),

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

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	taskList := []TaskData{}
	for _, t := range tasks {
		taskList = append(taskList, t)
	}

	data := map[string]any{
		"Title": "Go + HTMX task tracking",
		"Now":   time.Now().Format(time.RFC1123),
		"Tasks": taskList,
	}
	if err := a.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (a *App) timeSSE() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			a.hub.Broadcast(Event{
				Name: "time-update",
				Data: t.Format(time.RFC1123),
			})
		}
	}
}

func logRuntimeStats(every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()

	var m runtime.MemStats
	for range t.C {
		runtime.ReadMemStats(&m)

		log.Printf(
			"[runtime] open_conns=%d goroutines=%d heap_alloc=%dKB heap_inuse=%dKB sys=%dKB gc_cycles=%d",
			atomic.LoadInt64(&openConns),
			runtime.NumGoroutine(),
			m.HeapAlloc/1024,
			m.HeapInuse/1024,
			m.Sys/1024,
			m.NumGC,
		)
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
