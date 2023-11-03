package main

import (
	"Instance_of_Web_Server_API_in_Go/src/server"
	"fmt"
	"log"
	"net/http"
)

func main() {
	server := server.NewServer()

	http.HandleFunc("/new_ballot", server.HandleBallot)
	http.HandleFunc("/vote", server.HandleVote)
	http.HandleFunc("/result", server.HandleResult)

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
