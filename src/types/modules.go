//This file defines the types used in communication between the client and the server.
//The 3 requests are: ballot, vote, and result(request for the result of the ballot).

package modules

import "time"

// ------------------ Ballot ------------------
//NewBallotRequest is the request for creating a new ballot

type NewBallotRequest struct {
	Rule     string    `json:"rule"`
	Deadline time.Time `json:"deadline"`
	VoterIDs []string  `json:"voter-ids"`
	Alts     int       `json:"alts"`      //number of alternatives that can be chosen
	TieBreak []int     `json:"tie-break"` //if there is a tie, the alternative with the lowest index in this array wins
}

//NewBallotResponse is the response for creating a new ballot

type NewBallotResponse struct {
	BallotID string `json:"ballot-id"`
}

// ------------------ Vote ------------------
//VoteRequest is the request for voting
type VoteRequest struct {
	AgentID  string `json:"agent-id"`
	BallotID string `json:"ballot-id"`
	Prefs    []int  `json:"prefs"`
	Options  []int  `json:"options,omitempty"` //Options is marked with omiteempty because right now I don't know its purpose
}

//VoteResponse is the return code for voting. Unnecessary to define a type for this.

// ------------------ Result ------------------
//ResultRequest is the request for the result of the ballot
type ResultRequest struct {
	BallotID string `json:"ballot-id"`
}

//ResultResponse is the response for the result of the ballot
type ResultResponse struct {
	Winner  int   `json:"winner"`
	Ranking []int `json:"ranking"`
}
