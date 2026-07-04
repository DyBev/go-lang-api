package main

import (
	"time"
)

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

