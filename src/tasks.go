package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

type TaskData struct {
	TaskID int
	TaskName string
	Completed bool
}

var tasks map[string]TaskData = make(map[string]TaskData)

func (a *App) registerTask(w http.ResponseWriter, r *http.Request) {
	a.taskID += 1
	taskName := r.FormValue("taskName")
	tasks[strconv.Itoa(a.taskID)] = TaskData{
		TaskID: a.taskID,
		TaskName: taskName,
		Completed: false,
	}

	w.WriteHeader(http.StatusCreated)
	a.hub.Broadcast(Event{
		Name: "task-created",
		Data: strconv.Itoa(a.taskID),
	})
}

func (a *App) completeTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
	taskData, ok := tasks[taskID]
	log.Printf("completing task: %s, isValid: %t", taskID, ok)
	if !ok {
		io.WriteString(w, "")
		return
	}

	taskData.Completed = true
	tasks[taskID] = taskData

	a.hub.Broadcast(Event{
		Name: "task-completed",
		Data: taskID,
	})
}

