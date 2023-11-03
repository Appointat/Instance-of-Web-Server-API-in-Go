package server

import (
	methods "Instance_of_Web_Server_API_in_Go/src/methods"
	modules "Instance_of_Web_Server_API_in_Go/src/types"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// Ballot stores the information of a ballot
type Ballot struct {
	ID       string
	Rule     string
	Deadline time.Time
	voterIDs []string
	Votes    map[string][]int
	Alts     int
	TieBreak []int
	Winner   int
}

// method for ballot
type Server struct {
	ballots      map[string]Ballot
	numBallots   int //the number of ballots
	BallotNextID int //the next ballot ID to be assigned
	validRules   []string
}

// methods for server
func NewServer() *Server {
	return &Server{make(map[string]Ballot), 0, 0, []string{"Majority", "Borda", "Condorcet"}}
}

func (server *Server) HandleBallot(w http.ResponseWriter, r *http.Request) {
	//Analyse the request
	var req modules.NewBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the rule is valid, here only Majority, Borda, Condorcet are allowed
	valid := false
	for _, rule := range server.validRules {
		if req.Rule == rule {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "Invalid rule", http.StatusBadRequest)
		return
	}

	//check if the deadline is valid
	if req.Deadline.Before(time.Now()) {
		http.Error(w, "Invalid deadline", http.StatusBadRequest)
		return
	}

	//check if the number of alternatives is valid and if the tie-break array is valid
	if req.Alts < 2 {
		http.Error(w, "Invalid number of alternatives", http.StatusBadRequest)
		return
	}
	if len(req.TieBreak) != req.Alts {
		http.Error(w, "Invalid tie-break array", http.StatusBadRequest)
		return
	}

	//create a map to keep track of the seen alternatives
	seen := make(map[int]bool)
	for _, alt := range req.TieBreak {
		if alt <= 0 || alt > req.Alts {
			http.Error(w, "Invalid alternative in the tie-break array", http.StatusBadRequest)
			return
		}
		if seen[alt] {
			http.Error(w, "Duplicate alternative in the tie-break array", http.StatusBadRequest)
			return
		}
		seen[alt] = true
	}
	//check if all the alternatives are covered
	if len(seen) != req.Alts {
		http.Error(w, "Not all alternatives are covered in tie-break array", http.StatusBadRequest)
		return
	}

	//convert server.numBallots to string to be the ballot ID
	ballotID := fmt.Sprint(req.Rule, server.numBallots)

	server.numBallots++

	//create a new ballot
	ballot := Ballot{
		ID:       ballotID,
		Rule:     req.Rule,
		Deadline: req.Deadline,
		voterIDs: req.VoterIDs,
		Votes:    make(map[string][]int),
		Alts:     req.Alts,
		TieBreak: req.TieBreak,
		Winner:   -1,
	}
	server.ballots[ballotID] = ballot

	//return the ballot ID
	response := modules.NewBallotResponse{
		BallotID: ballotID,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (server *Server) HandleVote(w http.ResponseWriter, r *http.Request) {
	var req modules.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the ballot ID is valid
	ballot, ok := server.ballots[req.BallotID]
	if !ok {
		http.Error(w, "ballot ID not found", http.StatusBadRequest)
		return
	}

	//check if the voter ID is valid
	validVoterID := false
	for _, voterID := range ballot.voterIDs {
		if voterID == req.AgentID {
			validVoterID = true
			break
		}
	}
	if !validVoterID {
		http.Error(w, "this voter ID is not allowed to vote", http.StatusBadRequest)
		return
	}

	//check the time now is still within the deadline,if not return 503
	if ballot.Deadline.Before(time.Now()) {
		http.Error(w, "too late", http.StatusServiceUnavailable)
		return
	}

	//check if the number of preferences is valid
	if len(req.Prefs) != ballot.Alts {
		http.Error(w, "invalid number of preferences", http.StatusBadRequest)
		return
	}
	//check if the preferences are valid
	seen := make(map[int]bool)
	for _, pref := range req.Prefs {
		if pref <= 0 || pref > ballot.Alts {
			http.Error(w, "invalid preference", http.StatusBadRequest)
			return
		}
		if seen[pref] {
			http.Error(w, "duplicate preference", http.StatusBadRequest)
			return
		}
		seen[pref] = true
	}
	//check if all the alternatives are covered
	if len(seen) != ballot.Alts {
		http.Error(w, "not all alternatives are covered", http.StatusBadRequest)
		return
	}

	//check if the vote has already been casted, if so return 403
	_, ok = ballot.Votes[req.AgentID]
	if ok {
		http.Error(w, "vote already casted", http.StatusForbidden)
		return
	}

	//cast the vote
	server.ballots[req.BallotID].Votes[req.AgentID] = req.Prefs

	//return 200
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vote pris en compte"))
}

func (server *Server) HandleResult(w http.ResponseWriter, r *http.Request) {
	var req modules.ResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the ballot ID is valid, return 404 if not
	ballot, ok := server.ballots[req.BallotID]
	if !ok {
		http.Error(w, "ballot ID not found", http.StatusNotFound)
		return
	}

	//check if the result is ready, if not return 425
	//if there are no votes, the result is not ready
	if len(ballot.Votes) == 0 {
		http.Error(w, "result not ready", 425)
		return
	}
	//Calculate the result
	// validRules := []string{"Majority", "Borda", "Condorcet"}
	var response modules.ResultResponse
	rankings := make([]int, 0)
	switch ballot.Rule {
	case "Majority":
		var prefs methods.Profile
		for voterID := range ballot.Votes {
			prefs = append(prefs, ballot.Votes[voterID])
		}
		var winners []int
		winners, _ = methods.MajoritySCF(prefs)
		ballot := server.ballots[req.BallotID]
		ballot.Winner = winners[0]
		server.ballots[req.BallotID] = ballot

		candidate_with_ranking, _ := methods.MajoritySWF(prefs)

		type CandidateRankingPair struct {
			Candidate int
			Ranking   int
		}
		var pairs []CandidateRankingPair
		for candidate, ranking := range candidate_with_ranking {
			pairs = append(pairs, CandidateRankingPair{Candidate: candidate, Ranking: ranking})
		}

		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Ranking < pairs[j].Ranking
		})

		rankings = []int{}
		for _, pair := range pairs {
			rankings = append(rankings, pair.Candidate)
		}

		for i := 0; i < len(rankings); i++ {
			fmt.Println(rankings[i])
		}

		// case "Borda":
		// 	methods.Borda()
		// case "Condorcet":
		// 	methods.Condorcet()
	}

	ballot = server.ballots[req.BallotID]
	if ballot.Winner == -1 {
		response = modules.ResultResponse{}
	} else {
		response = modules.ResultResponse{
			Winner:  ballot.Winner,
			Ranking: rankings,
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
