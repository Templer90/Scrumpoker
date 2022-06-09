package models

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"
)

type sessionData struct {
	Session     *ScrumpokerSession
	LastUpdated time.Time
}

type SessionManager struct {
	sessions           map[string]sessionData
	maxSessionLifetime time.Duration
}

// Note - NOT RFC4122 compliant
func genUuid() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func InitilaizeManager(maxSessionLifetime time.Duration) *SessionManager {
	sessionmg := new(SessionManager)
	sessionmg.sessions = make(map[string]sessionData)
	sessionmg.maxSessionLifetime = maxSessionLifetime

	return sessionmg
}

func (sm SessionManager) AddSession(admin User, cards []string) (*ScrumpokerSession, error) {
	session, err := InitilaizeScrumpoker(admin, cards)
	if err != nil {
		return nil, err
	}

	sm.sessions[session.SessionID] = sessionData{Session: session, LastUpdated: time.Now()}

	return session, nil
}

func (sm SessionManager) JoinSession(SessionID string, user User) (*ScrumpokerSession, error) {
	sessionData, err := sm.findSessionData(SessionID)
	if err != nil {
		return nil, err
	}

	sessionData.Session.Users = append(sessionData.Session.Users, user)
	sessionData.LastUpdated = time.Now()

	return sessionData.Session, nil
}

func (sm SessionManager) DeleteSession(SessionID string, admin User) error {
	sessionData, err := sm.findSessionData(SessionID)
	if err != nil {
		return err
	}

	if sessionData.Session.Creator.UUID != admin.UUID {
		return errors.New("not Admin")
	}

	delete(sm.sessions, SessionID)
	return nil
}

func (sm SessionManager) GetSession(SessionID string) (*ScrumpokerSession, error) {
	sessionData, err := sm.findSessionData(SessionID)
	if err != nil {
		return nil, err
	}

	sessionData.LastUpdated = time.Now()
	return sessionData.Session, nil
}

func (sm SessionManager) findSessionData(SessionID string) (*sessionData, error) {
	sessionData, err := sm.sessions[SessionID]
	if !err {
		return nil, errors.New("no Session Found")
	}

	return &sessionData, nil
}

func (sm SessionManager) Cleanup() {
	now := time.Now()
	oldSessions := make([]string, 1)

	for _, session := range sm.sessions {
		if now.Sub(session.LastUpdated).Seconds() > sm.maxSessionLifetime.Seconds() {
			oldSessions = append(oldSessions, session.Session.SessionID)
		}
	}

	for _, sessionID := range oldSessions {
		delete(sm.sessions, sessionID)
	}
}
