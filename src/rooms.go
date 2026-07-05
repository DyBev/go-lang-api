package main

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RoomData struct {
	Clients []string
	RoomName string
	RoomCode string
}

var rooms map[string]RoomData = make(map[string]RoomData)
var joinCodes map[string]string = make(map[string]string)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomChar() (byte, error) {
	max := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func generateCode() (string, error) {
	code := make([]byte, 9)

	for i := range 9 {
		if i == 4 {
			code[i] = '-'
			continue
		}

		c, err := randomChar()
		if err != nil {
			return "", err
		}
		code[i] = c
	}

	return string(code), nil
}

func (a *App) createRoom(w http.ResponseWriter, r *http.Request) {
	roomName := r.FormValue("room-name")
	clientID := r.Context().Value(clientIDKey).(string)
	roomID := uuid.NewString()
	for _, ok := rooms[roomID]; ok; {
		roomID = uuid.NewString()
	}

	joinCode, err := generateCode()
	if err != nil {
		panic(err)
	}

	for _, ok := joinCodes[joinCode]; ok; {
		joinCode, err = generateCode()
		if err != nil {
			panic(err)
		}
	}

	rooms[roomID] = RoomData{
		Clients: []string{clientID},
		RoomName: roomName,
		RoomCode: joinCode,
	}
	joinCodes[joinCode] = roomID

	w.WriteHeader(http.StatusCreated)
	data := map[string]any{
		"Title": "Go + HTMX task tracking",
		"Now":   time.Now().Format(time.RFC1123),
		"RoomCode": joinCode,
		"RoomName": roomName,
		"RoomID": roomID,
	}
	if err := a.templates.ExecuteTemplate(w, "room-content.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func (a *App) joinRoom(w http.ResponseWriter, r *http.Request) {
	roomCode := r.FormValue("room-code")
	clientID := r.Context().Value(clientIDKey).(string)
	roomID, ok := joinCodes[roomCode]
	if !ok {
		data := map[string]any{
			"ErrorMessage": "Room code invalid, please try again",
		}
		if err := a.templates.ExecuteTemplate(w, "join-content-error.html", data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	}

	roomData := rooms[roomID]
	roomData.Clients = append(roomData.Clients, clientID)

	rooms[roomID] = roomData

	taskList := []TaskData{}
	for _, t := range tasks {
		if t.Room == roomID {
			taskList = append(taskList, t)
		}
	}

	data := map[string]any{
		"Title": "Go + HTMX task tracking",
		"Now":   time.Now().Format(time.RFC1123),
		"RoomCode": roomData.RoomCode,
		"RoomName": roomData.RoomName,
		"RoomID": roomID,
		"Tasks": taskList,
	}
	if err := a.templates.ExecuteTemplate(w, "room-content.html", data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}
