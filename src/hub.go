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

type Client struct {
	Events chan Event
}

type Hub struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
	rooms map[string]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe(roomID string) chan Event {
	ch := make(chan Event, 10)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms == nil {
		h.rooms = make(map[string]map[chan Event]struct{})
	}

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[chan Event]struct{})
	}

	h.rooms[roomID][ch] = struct{}{}

	return ch
}

func (h *Hub) Unsubscribe(roomID string, ch chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms == nil {
		return
	}

	if _, ok := h.rooms[roomID]; !ok {
		return
	}

	delete(h.rooms[roomID], ch)
	close(ch)
}

func (h *Hub) Broadcast(roomID string, ev Event) {
	h.mu.RLock()
	subs := h.rooms[roomID]
	h.mu.RUnlock()

	for ch := range subs {
		select {
		case ch <- ev:
		default:
			// drop slow clients
		}
	}
}

func (a *App) eventsSSE(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	ch := a.hub.Subscribe(roomID)
	defer a.hub.Unsubscribe(roomID, ch)

	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return

		case ev := <-ch:
			switch ev.Name {

			case "task-created":
				taskID := ev.Data

				taskData, ok := tasks[taskID]
				if !ok {
					continue
				}

				data := map[string]string{
					"TaskID":   taskID,
					"TaskName": taskData.TaskName,
				}

				var b bytes.Buffer
				if err := a.templates.ExecuteTemplate(&b, "task-uncompleted.html", data); err != nil {
					continue
				}

				html := strings.ReplaceAll(b.String(), "\n", "")
				fmt.Fprint(w, "event: task-created\n")
				fmt.Fprintf(w, "data: %s\n\n", html)

			case "task-completed":
				taskID := ev.Data

				taskData, ok := tasks[taskID]
				if !ok {
					continue
				}

				data := map[string]string{
					"TaskName": taskData.TaskName,
				}

				var b bytes.Buffer
				if err := a.templates.ExecuteTemplate(&b, "task-completed.html", data); err != nil {
					continue
				}

				html := strings.ReplaceAll(b.String(), "\n", "")
				fmt.Fprintf(w, "event: task-%s-completed\n", taskID)
				fmt.Fprintf(w, "data: %s\n\n", html)

			case "time-update":
				fmt.Fprintf(w, "event: time-update\n")
				fmt.Fprintf(w, "data: %s\n\n", ev.Data)
			}

			flusher.Flush()
		}
	}
}
