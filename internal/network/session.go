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

var sessionFilePath = "assets/data/sessions.json"

// ReadSessions đọc tất cả các session từ file JSON
func ReadSessions() ([]Session, error) {
	log.Println("[SESSION] Reading all sessions")

	var sessions []Session
	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		log.Println("[SESSION] Session file not found. Returning empty list.")
		return sessions, nil
	}

	file, err := os.Open(sessionFilePath)
	if err != nil {
		log.Println("[SESSION] Failed to open session file:", err)
		return sessions, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err == io.EOF {
		log.Println("[SESSION] Session file is empty.")
		return sessions, nil
	} else if err != nil {
		log.Println("[SESSION] Error decoding session file:", err)
		return sessions, err
	}

	log.Printf("[SESSION] Loaded %d sessions from file", len(sessions))
	return sessions, nil
}

// ReadSession reads the session data from the file
func ReadSession(sessionID string) (Session, error) {
	log.Printf("[SESSION] Reading session with ID: %s", sessionID)

	var sessions []Session
	var session Session

	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		log.Println("[SESSION] Session file does not exist")
		return Session{Authenticated: false, Username: "", SessionID: ""}, nil
	}

	file, err := os.Open(sessionFilePath)
	if err != nil {
		log.Println("[SESSION] Failed to open session file:", err)
		return session, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err == io.EOF {
		log.Println("[SESSION] No sessions in file")
		return session, fmt.Errorf("session with id %s not found", sessionID)
	} else if err != nil {
		log.Println("[SESSION] Failed to decode session file:", err)
		return session, err
	}

	for _, s := range sessions {
		if s.SessionID == sessionID {
			log.Printf("[SESSION] Found session for user %s", s.Username)
			return s, nil
		}
	}

	log.Printf("[SESSION] Session with ID %s not found", sessionID)
	return session, fmt.Errorf("session with id %s not found", sessionID)
}

// WriteSession writes the session data to the file
func WriteSession(sessions []Session) error {
	log.Println("[SESSION] Writing sessions to file")

	latest := make(map[string]Session)
	for _, s := range sessions {
		latest[s.Username] = s
	}

	var uniqueSessions []Session
	for _, s := range latest {
		uniqueSessions = append(uniqueSessions, s)
	}

	file, err := os.OpenFile(sessionFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("[SESSION] Failed to open file for writing:", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(uniqueSessions)
	if err != nil {
		log.Println("[SESSION] Failed to encode sessions to file:", err)
		return err
	}

	log.Printf("[SESSION] Successfully wrote %d unique sessions", len(uniqueSessions))
	return nil
}

func FindSessionByID(sessionID string) (Session, error) {
	log.Printf("[SESSION] Finding session by ID: %s", sessionID)

	file, err := os.Open(sessionFilePath)
	if err != nil {
		log.Println("[SESSION] Could not open session file:", err)
		return Session{}, err
	}
	defer file.Close()

	var sessions []Session
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err != nil {
		log.Println("[SESSION] Could not decode session file:", err)
		return Session{}, err
	}

	for _, s := range sessions {
		if s.SessionID == sessionID {
			log.Printf("[SESSION] Session found: user=%s", s.Username)
			return s, nil
		}
	}

	log.Printf("[SESSION] Session ID %s not found", sessionID)
	return Session{}, fmt.Errorf("session not found")
}
