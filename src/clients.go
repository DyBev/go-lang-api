package main

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type key string
const clientIDKeystring string = "client_id"
const clientIDKey key = "client_id"

type ClientData struct {
	lastSeen time.Time
}

var clients map[string]ClientData = make(map[string]ClientData)

func clientIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(clientIDKeystring)
		var id string

		if err != nil || cookie.Value == "" {
			id = uuid.NewString()
			clients[id] = ClientData{
				lastSeen: time.Now(),
			}

			http.SetCookie(w, &http.Cookie{
				Name:     clientIDKeystring,
				Value:    id,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		} else {
			id = cookie.Value
			clients[id] = ClientData{
				lastSeen: time.Now(),
			}
		}

		// store in request context
		ctx := context.WithValue(r.Context(), clientIDKey, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) clientHeartbeat(w http.ResponseWriter, r *http.Request) {
	clientID := r.Context().Value(clientIDKey).(string)

	clients[clientID] = ClientData{
		lastSeen: time.Now(),
	}
	w.WriteHeader(http.StatusNoContent)
}

func cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)

	for range ticker.C {
		now := time.Now()

		for clientID, client := range clients {
			if now.Sub(client.lastSeen) > 60*time.Second {
				delete(clients, clientID)
			}
		}
	}
}
