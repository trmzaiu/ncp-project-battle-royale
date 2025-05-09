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



// ReadSession reads the session data from the file
func ReadSession(sessionID string) (Session, error) {
	var sessions []Session
	var session Session

	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		// If session file does not exist, create a default session and return
		session = Session{Authenticated: false, Username: "", SessionID: ""}
		err := WriteSession([]Session{session})
		if err != nil {
			return session, err
		}
		log.Println("Session file created with default session.")
		return session, nil
	}

	// Read the session file
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return session, err
	}
	defer file.Close()

	// Decode the sessions into a slice
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessions)
	if err == io.EOF {
		return session, nil
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
	return session, fmt.Errorf("session with id %s not found", sessionID)
}

// WriteSession writes the session data to the file
func WriteSession(sessions []Session) error {
	file, err := os.Create(sessionFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode and write sessions to the file
	encoder := json.NewEncoder(file)
	return encoder.Encode(sessions)
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