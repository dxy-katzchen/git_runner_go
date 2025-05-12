package main

import (
	"log"
	"net/http"
	"git-runner/handler"
)

func main() {
	http.HandleFunc("/webhook", handler.WebhookHandler)
	log.Println("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
