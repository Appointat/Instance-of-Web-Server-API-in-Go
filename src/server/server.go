package server

import (
	modules "Instance_of_Web_Server_API_in_Go/src/types"
	"encoding/json"
	"fmt"
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

//method for ballot

// rank returns the index of the alternative in one's preference list
func (b Ballot) rank(alt int, VoterID string) int {
	for i, pref := range b.Votes[VoterID] {
		if pref == alt {
			return i
		}
	}
	return -1
}

// isPref func returns true if alt1 is preferred to alt2 by the voter with the given ID
func (b Ballot) isPref(alt1, alt2 int, VoterID string) bool {
	return b.rank(alt1, VoterID) < b.rank(alt2, VoterID)
}

func (b *Ballot) Majority() {
	//count the number of votes for each alternative
	count := make([]int, b.Alts)
	for _, prefs := range b.Votes {
		count[prefs[0]]++
	}
	//find the alternative with the most votes, if there is a tie, use the tie-break array.
	//The alternative with the lowest index in the tie-break array wins
	max := 0
	for i := 1; i < b.Alts; i++ {
		if count[i] > count[max] || (count[i] == count[max] && b.TieBreak[i] < b.TieBreak[max]) {
			max = i
		}
	}
	b.Winner = max
}

func (b *Ballot) Borda() {
	//Borda rule: the alternative with the highest score wins
	score := make([]int, b.Alts)
	for _, prefs := range b.Votes {
		for i, pref := range prefs {
			score[pref] += b.Alts - i
		}
	}
	//find the alternative with the highest score, if there is a tie, use the tie-break array.
	max := 0
	for i := 1; i < b.Alts; i++ {
		if score[i] > score[max] || (score[i] == score[max] && b.TieBreak[i] < b.TieBreak[max]) {
			max = i
		}
	}
}

func (b *Ballot) Condorcet() {
	// Condorcet rule: the alternative that wins in all the pairwise comparisons wins
	// create a matrix to store the pairwise comparison results
	matrix := make([][]int, b.Alts)
	for i := 0; i < b.Alts; i++ {
		matrix[i] = make([]int, b.Alts)
	}

	// fill in the matrix
	for i := 0; i < b.Alts; i++ {
		for j := 0; j < b.Alts; j++ {
			if i != j {
				for _, voterID := range b.VoterIDs {
					if b.isPref(i, j, voterID) {
						matrix[i][j]++
					}
				}
			}
		}
	}

	// find the alternatives that win in all the pairwise comparisons
	winners := []int{}
	for i := 0; i < b.Alts; i++ {
		win := true
		for j := 0; j < b.Alts; j++ {
			if i != j && matrix[i][j] <= matrix[j][i] {
				win = false
				break
			}
		}
		if win {
			winners = append(winners, i)
		}
	}

	// handle cases
	switch len(winners) {
	case 0:
		// no Condorcet winner
		b.Winner = -1
	case 1:
		// a single Condorcet winner
		b.Winner = winners[0]
	default:
		// multiple Condorcet winners, use the TieBreak array
		b.Winner = winners[0]
		for _, alt := range b.TieBreak {
			for _, winner := range winners {
				if winner == alt {
					b.Winner = winner
					return
				}
			}
		}
	}
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

func (s *Server) HandleBallot(w http.ResponseWriter, r *http.Request) {
	//Analyse the request
	var req modules.NewBallotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the rule is valid, here only Majority, Borda, Condorcet are allowed
	validRules := []string{"Majority", "Borda", "Condorcet"}
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

	//affection of the ballot ID, the ballot ID is "scrutin" + the next ballot ID(from number to string)
	ballotID := fmt.Sprintf("scrutin%d", s.BallotNextID)

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

func (s *Server) HandleVote(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) HandleResult(w http.ResponseWriter, r *http.Request) {
	var req modules.ResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	//check if the ballot ID is valid, return 404 if not
	ballot, ok := s.Ballots[req.BallotID]
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
	switch ballot.Rule {
	case "Majority":
		ballot.Majority()
	case "Borda":
		ballot.Borda()
	case "Condorcet":
		ballot.Condorcet()
	}
	s.Ballots[req.BallotID] = ballot
	var response modules.ResultResponse
	if ballot.Winner == -1 {
		response = modules.ResultResponse{}
	} else {
		response = modules.ResultResponse{
			Winner:  ballot.Winner,
			Ranking: []int{},
		}
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
