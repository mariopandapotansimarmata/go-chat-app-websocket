package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	setupAPI()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupAPI() {
	ctx := context.Background()
	manager := NewManager(ctx)
	http.HandleFunc("/login", manager.LoginHandler)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serveWS)
}
