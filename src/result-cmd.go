package main

import (
	"encoding/json"
	"net/http"
)

// Structure for the request payload
type ResultRequest struct {
	BallotID string `json:"ballot-id"`
}

// Structure for the response payload
type ResultResponse struct {
	Winner  int   `json:"winner,omitempty"`
	Ranking []int `json:"ranking,omitempty"`
}

func handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var request ResultRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Mock judgement for ballot-id. This is just for demonstration,
	// in reality, you might need to query a database or do other operations
	if request.BallotID == "scrutin12" {
		// If the ballot-id is "scrutin12", return the sample data
		response := ResultResponse{
			Winner:  4,
			Ranking: []int{2, 1, 4, 3},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else if request.BallotID == "" {
		http.Error(w, "Too early", 425)
	} else {
		http.Error(w, "Not Found", 404)
	}
}

func main() {
	http.HandleFunc("/result", handleResult)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
