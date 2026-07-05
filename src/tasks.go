package main

import (
	"io"
	"net/http"
	"strconv"
)

type TaskData struct {
	TaskID int
	TaskName string
	Completed bool
	Room string
}

var tasks map[string]TaskData = make(map[string]TaskData)

func (a *App) registerTask(w http.ResponseWriter, r *http.Request) {
	a.taskID += 1
	taskName := r.FormValue("taskName")
	roomID := r.FormValue("roomID")
	tasks[strconv.Itoa(a.taskID)] = TaskData{
		TaskID: a.taskID,
		TaskName: taskName,
		Completed: false,
		Room: roomID,
	}

	w.WriteHeader(http.StatusCreated)
	a.hub.Broadcast(roomID, Event{
		Name: "task-created",
		Data: strconv.Itoa(a.taskID),
	})
}

func (a *App) completeTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("taskID")
	roomID := r.FormValue("roomID")

	taskData, ok := tasks[taskID]
	if !ok {
		io.WriteString(w, "")
		return
	}

	taskData.Completed = true
	tasks[taskID] = taskData

	a.hub.Broadcast(roomID, Event{
		Name: "task-completed",
		Data: taskID,
	})
}

