// internal/network/session.go

package network

import (
	"encoding/json"
	"fmt"
	"os"
)

// Session store
type Session struct {
	SessionID     string `json:"session_id"`
	Username      string `json:"username"`
	Authenticated bool   `json:"authenticated"`
}

// File to store session data
var sessionFilePath = "assets/data/sessions.json"

// ReadSession reads the session data from the file
func ReadSessions() ([]Session, error) {
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sessions []Session
	err = json.NewDecoder(file).Decode(&sessions)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// WriteSession writes the session data to the file
func WriteSession(newSession Session) error {
	sessions, err := ReadSessions()
	if err != nil {
		sessions = []Session{}
	}

	updated := false
	for i, s := range sessions {
		if s.Username == newSession.Username {
			sessions[i] = newSession
			updated = true
			break
		}
	}
	if !updated {
		sessions = append(sessions, newSession)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionFilePath, data, 0644)
}

func FindSessionByID(sessionID string) (Session, error) {
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	var session Session
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return Session{}, err
	}

	if session.SessionID == sessionID {
		return session, nil
	}
	return Session{}, fmt.Errorf("session not found")
}