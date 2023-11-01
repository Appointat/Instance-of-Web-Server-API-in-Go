package main

import (
	"encoding/json"
	"net/http"
)

type VoteRequest struct {
	AgentID  string `json:"agent-id"`
	BallotID string `json:"ballot-id"`
	Prefs    []int  `json:"prefs"`
	Options  []int  `json:"options,omitempty"`
}

func handleVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var request VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Add your logic for processing the vote here

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vote pris en compte"))
}

func main() {
	http.HandleFunc("/vote", handleVote)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
