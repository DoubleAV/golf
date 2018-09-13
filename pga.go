package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PGA struct {
	lastUpdated time.Time
	leaderboard *Leaderboard
	tid         string
}

//Don't really understand why he has this here, just returns a given string.
func (pga *PGA) String() string {
	return "PGA Tour"
}

//Returns the TID of the PGA Tour
func (pga *PGA) TID() string {
	return pga.tid
}

//Function to update tournament ID
func (pga *PGA) UpdateTID() error {
	var current struct {
		TID string `json:"tid"`
	}

	//send request to get current tournament info
	resp, err := client.Get("https://statdata.pgatour.com/r/current/message.json")
	//If there is an error, return it
	if err != nil {
		return err
	}
	//close response body when finished
	defer resp.Body.Close()
	//decode error if there is one, and return it
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		return err
	}
	//Throw error if no tournament ID is returned
	if current.TID == "" {
		return errors.New("TID is empty")
	}
	//Set tournament ID
	pga.tid = current.TID
	return nil
}

//Request() makes a request to the PGATour leaderboard with a certain tournament ID and returns
//The Request
func (pga *PGA) Request() (*http.Request, error) {
	//Sprintf formats string without printing it.
	//In this case, pga.TID is inserted where %s is.
	//
	u := fmt.Sprintf("https://statdata.pgatour.com/r/%s/leaderboard-v2mini.json", pga.TID())
	return http.NewRequest("GET", u, nil)
}

func (pga *PGA) Parse(r io.Reader) (*Leaderboard, error) {
	var d PGALeaderboard
	if err := json.NewDecoder(r).Decode(&d); err != nil {
		return nil, err
	}

	var players []*Player

	for _, p := range d.Leaderboard.Players {
		var rounds []int
		for _, r := range p.Rounds {
			rounds = append(rounds, r.Strokes)
		}
		players = append(players, &Player{
			Name:            p.PlayerBio.FirstName + " " + p.PlayerBio.LastName,
			Country:         p.PlayerBio.Country,
			CurrentPosition: p.CurrentPosition,
			StartPosition:   p.StartPosition,
			Today:           p.Today,
			Total:           p.Total,
			After:           p.Thru,
			Hole:            p.CourseHole,
			TotalStrokes:    p.TotalStrokes,
			Rounds:          rounds,
		})
	}

	return &Leaderboard{
		Tour:       pga.String(),
		Tournament: d.Leaderboard.TournamentName,
		Course:     d.Leaderboard.Courses[0].CourseName,
		Date:       fmt.Sprintf("%s — %s", d.Leaderboard.StartDate, d.Leaderboard.EndDate),
		Players:    players,
		Updated:    d.LastUpdated,
		Round:      d.Leaderboard.CurrentRound,
	}, nil
}

//Sets PGA Leaderboard
func (pga *PGA) SetLeaderboard(lb *Leaderboard) {
	pga.leaderboard = lb
}

//returns PGA Leaderboard
func (pga *PGA) Leaderboard() *Leaderboard {
	return pga.leaderboard
}

//Returns time the PGA was last updated
func (pga *PGA) LastUpdated() time.Time {
	return pga.lastUpdated
}

//returns the current time
func (pga *PGA) SetLastUpdated(t time.Time) {
	pga.lastUpdated = t
}

type PGALeaderboard struct {
	LastUpdated string `json:"last_updated"`
	Leaderboard struct {
		Courses []struct {
			CourseName string `json:"course_name"`
		}
		TournamentName string `json:"tournament_name"`
		TourName       string `json:"tour_name"`
		StartDate      string `json:"start_date"`
		EndDate        string `json:"end_date"`

		CurrentRound int `json:"current_round"`

		Players []struct {
			CourseHole      int    `json:"course_hole"`
			CurrentPosition string `json:"current_position"`
			StartPosition   string `json:"start_position"`
			Thru            int
			Today           int
			Total           int
			TotalStrokes    int `json:"total_strokes"`
			PlayerBio       struct {
				Country   string `json:"country"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				ShortName string `json:"short_name"`
			} `json:"player_bio"`
			Rounds []struct {
				RoundNumber int `json:"round_number"`
				Strokes     int `json:"strokes"`
			}
		} `json:"players"`
	} `json:"leaderboard"`
}
