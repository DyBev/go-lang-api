package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Name struct {
	Name string `json:"name"`
}

func main() {
	serveHello := func(w http.ResponseWriter, _ *http.Request) {
		log.Print("request at GET /hello")
		io.WriteString(w, "hello world")
	}

	serveBye := func(w http.ResponseWriter, _ *http.Request) {
		log.Print("request at GET /bye")
		io.WriteString(w, "bye world")
	}

	serveName := func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("error getting the body data %v", err)
			io.WriteString(w, "error getting body data")
			return
		}
		var bodyData Name
		if err := json.Unmarshal(body, &bodyData); err != nil {
			log.Printf("error unmarshalling the body data %v", err)
			io.WriteString(w, "error unmarshalling body data")
			return
		}

		log.Printf("request at POST /name body %s", bodyData.Name)
		io.WriteString(w, fmt.Sprintf("hello %s", bodyData.Name))
	}


	http.HandleFunc("/hello", serveHello)
	http.HandleFunc("/bye", serveBye)
	http.HandleFunc("POST /name", serveName)

	log.Print("listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
