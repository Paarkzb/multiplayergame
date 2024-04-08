package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	setupAPI()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// setupAPI will start all Routes and their Handlers
func setupAPI() {
	r := mux.NewRouter()
	r.HandleFunc("/ws", serveWS)
	http.Handle("/", r)
}
