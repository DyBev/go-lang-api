package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	serveHello := func(w http.ResponseWriter, _ *http.Request) {
		log.Print("request at GET /hello")
		io.WriteString(w, "hello world")
	}

	serveBye := func(w http.ResponseWriter, _ *http.Request) {
		log.Print("request at GET /bye")
		io.WriteString(w, "bye world")
	}

	http.HandleFunc("/hello", serveHello)
	http.HandleFunc("/bye", serveBye)

	log.Print("listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
