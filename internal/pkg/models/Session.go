package models

import (
	"errors"
	"strconv"
	"strings"
)

type ScrumpokerSession struct {
	SessionID string
	Cards     []string

	Creator User
	Users   []User
	Votes   map[User]string

	isNumeric  bool
	ShouldShow bool
}

type SessionStatus struct {
	Votes   []UserVote
	Average string `json:"Average"`
	Closest string `json:"Closest"`
}

type UserVote struct {
	Name string `json:"Name"`
	Vote string `json:"Vote"`
}

func InitilaizeScrumpoker(admin User, cards []string) (*ScrumpokerSession, error) {
	session := ScrumpokerSession{
		SessionID: genUuid(),
		Creator:   admin,
		Votes:     make(map[User]string),
	}
	session.Users = append(session.Users, admin)

	session.isNumeric = true
	filteredCards := make([]string, 0)
	for _, card := range cards {
		trimmedText := strings.TrimSpace(card)
		// No empty Cards
		if trimmedText == "" {
			continue
		}
		// The Cardtext could be a valid urlparameter
		if trimmedText == "status" {
			continue
		}
		//Determine if the Cards are numeric (and therefore the average can be calculated)
		_, err := strconv.ParseFloat(card, 64)
		if err != nil {
			session.isNumeric = false
		}
		filteredCards = append(filteredCards, card)
	}

	if len(filteredCards) < 2 {
		return nil, errors.New("at least 2 Cards needed")
	}
	session.Cards = filteredCards

	return &session, nil
}

func (session ScrumpokerSession) GetUser(uuid string) (*User, error) {
	for _, user := range session.Users {
		if user.UUID == uuid {
			return &user, nil
		}
	}

	if session.Creator.UUID == uuid {
		return &session.Creator, nil
	}
	return nil, errors.New("no User Found")
}

func (session ScrumpokerSession) GetAdmin(uuid string) (*User, error) {
	if session.Creator.UUID == uuid {
		return &session.Creator, nil
	}
	return nil, errors.New("no Admin Found")
}

func (session ScrumpokerSession) IsAdmin(uuid string) bool {
	return session.Creator.UUID == uuid
}

func (session *ScrumpokerSession) Vote(uuid string, vote string) {
	user, _ := session.GetUser(uuid)

	session.Votes[*user] = vote
}

func (session *ScrumpokerSession) ResetVotes() {
	for _, user := range session.Users {
		session.Votes[user] = ""
	}
}

func (session *ScrumpokerSession) AdministerSession(showCards bool) {
	session.ShouldShow = showCards
}

func (session ScrumpokerSession) Status(uuid string) SessionStatus {
	sum := 0.0
	status := SessionStatus{}
	for _, user := range session.Users {
		txt := ""
		if session.ShouldShow || user.UUID == uuid {
			txt = session.Votes[user]
		}
		vote := UserVote{
			Name: user.Name,
			Vote: txt,
		}
		status.Votes = append(status.Votes, vote)

		if session.isNumeric {
			v, _ := strconv.ParseFloat(session.Votes[user], 64)
			sum += v
		}
	}

	if session.isNumeric && session.ShouldShow {
		status.Average = strconv.FormatFloat(sum/(float64(len(session.Users))), 'f', 2, 64)
	}

	return status
}
