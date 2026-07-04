package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Event struct {
	Name string
	Data string
}

type Hub struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 16) // small buffer to avoid blocking
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- ev:
		default:
		}
	}
}

func (a *App) eventsSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	ch := a.hub.Subscribe()
	defer a.hub.Unsubscribe(ch)

	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			if ev.Name == "task-created" {
				taskID := ev.Data
				taskData, ok := tasks[taskID]
				if !ok {
					continue
				}

				data := map[string]string{
					"TaskID": taskID,
					"TaskName": taskData.name,
				}

				var b bytes.Buffer
				if err := a.templates.ExecuteTemplate(&b, "task-uncompleted.html", data); err != nil {
					continue
				}

				html := strings.ReplaceAll(b.String(), "\n", "")
				fmt.Fprint(w, "event: task-created\n")
				fmt.Fprintf(w, "data: %s\n\n", html)
			}

			if ev.Name == "task-completed" {
				taskID := ev.Data
				taskData, ok := tasks[taskID]
				if !ok {
					continue
				}

				data := map[string]string{
					"TaskName": taskData.name,
				}

				var b bytes.Buffer
				if err := a.templates.ExecuteTemplate(&b, "task-completed.html", data); err != nil {
					fmt.Fprintf(w, "event: task-%s-completed\n", ev.Name)
					fmt.Fprint(w, "data: \"\"\n\n")
					continue
				}

				html := strings.ReplaceAll(b.String(), "\n", "")
				fmt.Fprintf(w, "event: task-%s-completed\n", ev.Data)
				fmt.Fprintf(w, "data: %s\n\n", html)
			}

			if ev.Name == "time-update" {
				fmt.Fprintf(w, "event: %s\n", ev.Name)
				fmt.Fprintf(w, "data: %s\n\n", ev.Data)
			}

			flusher.Flush()
		}
	}
}
