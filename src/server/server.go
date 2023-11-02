package server

import (
	modules "Instance_of_Web_Server_API_in_Go/src/types"
	"encoding/json"
	"net/http"
	"time"
)

// Ballot stores the information of a ballot
type Ballot struct {
	ID       string
	Rule     string
	Deadline time.Time
	VoterIDs []string
	Votes    map[string][]int
	Alts     int
	TieBreak []int
	Winner   int
}

type Server struct {
	Ballots      map[string]Ballot
	BallotNextID int //the next ballot ID to be assigned
	NumBallots   int //the number of ballots
}

// methods for server
func NewServer() *Server {
	return &Server{make(map[string]Ballot), 0, 0}
}

func (s *Server) handleBallot(w http.ResponseWriter, r *http.Request) {
	//Analyse the request
	var req modules.NewBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the rule is valid, here only Majority、Approval、Condorcet are allowed
	validRules := []string{"Majority", "Approval", "Condorcet"}
	valid := false
	for _, rule := range validRules {
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
		if alt < 0 || alt >= req.Alts {
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

	//affection of the ballot ID
	ballotID := "scrutin" + string(s.BallotNextID)
	s.BallotNextID++
	s.NumBallots++

	//create a new ballot
	ballot := Ballot{
		ID:       ballotID,
		Rule:     req.Rule,
		Deadline: req.Deadline,
		VoterIDs: req.VoterIDs,
		Votes:    make(map[string][]int),
		Alts:     req.Alts,
		TieBreak: req.TieBreak,
		Winner:   -1,
	}
	s.Ballots[ballotID] = ballot

	//return the ballot ID
	response := modules.NewBallotResponse{
		BallotID: ballotID,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	var req modules.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the ballot ID is valid
	ballot, ok := s.Ballots[req.BallotID]
	if !ok {
		http.Error(w, "ballot ID not found", http.StatusBadRequest)
		return
	}

	//check if the voter ID is valid
	validVoterID := false
	for _, voterID := range ballot.VoterIDs {
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
		if pref < 0 || pref >= ballot.Alts {
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
	ballot.Votes[req.AgentID] = req.Prefs
	s.Ballots[req.BallotID] = ballot

	//return 200
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("vote pris en compte"))
}
