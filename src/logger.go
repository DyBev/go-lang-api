package main

import (
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

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

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
