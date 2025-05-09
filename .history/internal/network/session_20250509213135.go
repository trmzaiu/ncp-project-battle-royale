// internal/network/session.go

package network

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// ReadSessions đọc tất cả các session từ file JSON
func ReadSessions() ([]Session, error) {
	var sessions []Session

	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		return sessions, nil
	}

	file, err := os.Open(sessionFilePath)
	if err != nil {
		return sessions, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err == io.EOF {
		return sessions, nil
	} else if err != nil {
		return sessions, err
	}

	return sessions, nil
}

// ReadSession reads the session data from the file
func ReadSession(sessionID string) (Session, error) {
	var sessions []Session
	var session Session

	// Check if the session file exists
	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		// If session file does not exist, return an empty session
		log.Println("Session file does not exist, creating default session.")
		session = Session{Authenticated: false, Username: "", SessionID: ""}
		return session, nil
	}

	// Open the session file
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return session, err
	}
	defer file.Close()

	// Decode the sessions into a slice
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err == io.EOF {
		// No sessions found
		log.Println("No sessions found in file.")
		return session, fmt.Errorf("session with id %s not found", sessionID)
	} else if err != nil {
		return session, err
	}

	// Search for the session by session_id
	for _, s := range sessions {
		if s.SessionID == sessionID {
			return s, nil
		}
	}

	// Return an empty session if not found
	log.Printf("Session with ID %s not found", sessionID)
	return session, fmt.Errorf("session with id %s not found", sessionID)
}

// WriteSession writes the session data to the file
func WriteSession(sessions []Session) error {
	file, err := os.OpenFile(sessionFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode and write sessions to the file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(sessions)
	if err != nil {
		return err
	}

	return nil
}

func FindSessionByID(sessionID string) (Session, error) {
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	var sessions []Session
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err != nil {
		return Session{}, err
	}

	// Debug: Log sessions read from the file
	log.Printf("Sessions read from file: %+v", sessions)

	for _, s := range sessions {
		if s.SessionID == sessionID {
			return s, nil
		}
	}
	return Session{}, fmt.Errorf("session not found")
}
