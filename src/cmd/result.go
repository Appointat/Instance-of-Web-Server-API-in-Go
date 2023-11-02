package main

import (
	modules "Instance_of_Web_Server_API_in_Go/src/types"
	"encoding/json"
	"net/http"
)

func handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var request modules.ResultRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Add your logic for processing the result here
	// Mock judgement for ballot-id. This is just for demonstration,
	// in reality, you might need to query a database or do other operations
	if request.BallotID == "scrutin12" {
		// If the ballot-id is "scrutin12", return the sample data
		response := modules.ResultResponse{
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
