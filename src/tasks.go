package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

type TaskData struct {
	id int
	name string
	completed bool
}

var tasks map[string]TaskData = make(map[string]TaskData)

func (a *App) registerTask(w http.ResponseWriter, r *http.Request) {
	a.taskID += 1
	taskName := r.FormValue("taskName")
	data := map[string]any{
		"TaskID": a.taskID,
		"TaskName": taskName,
	}
	tasks[strconv.Itoa(a.taskID)] = TaskData{
		id: a.taskID,
		name: taskName,
		completed: false,
	}

	w.WriteHeader(http.StatusCreated)
	if err := a.templates.ExecuteTemplate(w, "task-uncompleted.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (a *App) completeTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
	taskData, ok := tasks[taskID]
	log.Printf("completing task: %s, isValid: %t", taskID, ok)
	if !ok {
		io.WriteString(w, "")
		return
	}

	taskData.completed = true
	tasks[taskID] = taskData

	data := map[string]string{
		"TaskName": taskData.name,
	}

	if err := a.templates.ExecuteTemplate(w, "task-completed.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

