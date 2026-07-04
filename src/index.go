package main

import (
	"net/http"
	"time"
)

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
