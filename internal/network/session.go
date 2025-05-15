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
	var sessions []Session
	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		log.Println("[WARN][SESSION] Session file not found")
		return sessions, nil
	}

	file, err := os.Open(sessionFilePath)
	if err != nil {
		log.Printf("[ERROR][SESSION] Failed to open file: %v", err)
		return sessions, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessions); err != nil {
		if err == io.EOF {
			log.Println("[INFO][SESSION] Session file is empty")
			return sessions, nil
		}
		log.Printf("[ERROR][SESSION] Failed to decode: %v", err)
		return sessions, err
	}

	return sessions, nil
}

// ReadSession reads the session data from the file

func ReadSession(sessionID string) (Session, error) {
	sessions, err := ReadSessions()
	if err != nil {
		return Session{}, err
	}

	for _, s := range sessions {
		if s.SessionID == sessionID {
			log.Printf("[INFO][SESSION] Session found for user: %s", s.Username)
			return s, nil
		}
	}

	log.Printf("[WARN][SESSION] Session ID %s not found", sessionID)
	return Session{}, fmt.Errorf("session with ID %s not found", sessionID)
}

// WriteSession writes the session data to the file
func WriteSession(sessions []Session) error {
	// Deduplicate by username
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
		log.Printf("[ERROR][SESSION] Failed to open file for writing: %v", err)
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(uniqueSessions); err != nil {
		log.Printf("[ERROR][SESSION] Failed to encode sessions: %v", err)
		return err
	}

	log.Printf("[INFO][SESSION] Wrote %d unique sessions", len(uniqueSessions))
	return nil
}

func FindSessionByID(sessionID string) (Session, error) {
	file, err := os.Open(sessionFilePath)
	if err != nil {
		log.Printf("[ERROR][SESSION] Could not open file: %v", err)
		return Session{}, err
	}
	defer file.Close()

	var sessions []Session
	if err := json.NewDecoder(file).Decode(&sessions); err != nil {
		log.Printf("[ERROR][SESSION] Could not decode file: %v", err)
		return Session{}, err
	}

	for _, s := range sessions {
		if s.SessionID == sessionID {
			log.Printf("[INFO][SESSION] Found user: %s", s.Username)
			return s, nil
		}
	}

	log.Printf("[WARN][SESSION] Session ID %s not found", sessionID)
	return Session{}, fmt.Errorf("session not found")
}
