// internal/network/server.go

package network

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"royaka/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
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
func ReadSession() (Session, error) {
	var session Session

	if _, err := os.Stat(sessionFilePath); os.IsNotExist(err) {
		session = Session{Authenticated: false, Username: "", SessionID: ""}
		err := WriteSession(session)
		if err != nil {
			return session, err
		}
		log.Println("Session file created with default session.")
		return session, nil
	}

	file, err := os.Open(sessionFilePath)
	if err != nil {
		return session, err
	}
	defer file.Close()

	fileStats, err := file.Stat()
	if err != nil {
		return session, err
	}
	if fileStats.Size() == 0 {
		session = Session{Authenticated: false, Username: "", SessionID: ""}
		return session, nil
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&session)
	if err == io.EOF {
		session = Session{Authenticated: false, Username: "", SessionID: ""}
		return session, nil
	} else if err != nil {
		return session, err
	}

	return session, nil
}

// WriteSession writes the session data to the file
func WriteSession(session Session) error {
	file, err := os.Create(sessionFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(session)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWS handles WebSocket connections and messages
func HandleWS(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			conn.Close()
			return
		}

		log.Printf("Received raw message: %s\n", string(msg))

		var pdu Message
		if err := json.Unmarshal(msg, &pdu); err != nil {
			log.Println("PDU decode error:", err)
			conn.WriteJSON(Response{Type: "error", Success: false, Message: "Invalid message format"})
			continue
		}

		// Handle different message types
		switch pdu.Type {
		case "register":
			var req RegisterRequest
			if err := json.Unmarshal(pdu.Data, &req); err != nil {
				log.Println("RegisterRequest decode error:", err)
				conn.WriteJSON(Response{Type: "register_response", Success: false, Message: "Invalid register data"})
				continue
			}

			// Hash the password before storing
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Password hashing error:", err)
				conn.WriteJSON(Response{Type: "register_response", Success: false, Message: "Error hashing password"})
				continue
			}

			// Add user to the database
			err = model.AddUser(*model.NewUser(req.Username, string(hashedPassword)))
			resp := Response{Type: "register_response", Success: err == nil, Message: "Registered successfully"}
			if err != nil {
				log.Println("Registration error:", err)
				resp.Message = "Registration failed: " + err.Error()
			}
			conn.WriteJSON(resp)

		case "login":
			var req LoginRequest
			if err := json.Unmarshal(pdu.Data, &req); err != nil {
				log.Println("LoginRequest decode error:", err)
				conn.WriteJSON(Response{Type: "login_response", Success: false, Message: "Invalid login data"})
				continue
			}

			// Find user and check password
			u, ok := model.FindUserByUsername(req.Username)
			resp := Response{Type: "login_response", Success: false, Message: "Invalid credentials"}
			if ok {
				// Compare the hashed password
				err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
				if err == nil {
					// Create a new session for the user
					sessionID := uuid.New().String()[:8]

					// Start a new session
					session := Session{
						SessionID:     sessionID,
						Username:      req.Username,
						Authenticated: true,
					}

					err := WriteSession(session)
					if err != nil {
						log.Println("Error writing session:", err)
						conn.WriteJSON(Response{Type: "login_response", Success: false, Message: "Error saving session"})
						continue
					}

					resp.Success = true
					resp.Message = "Login successful"

				} else {
					resp.Message = "Invalid credentials"
				}
			}
			conn.WriteJSON(resp)

		case "get_user":
			var req struct {
				SessionID string `json:"session_id"`
			}

			if err := json.Unmarshal(pdu.Data, &req); err != nil {
				conn.WriteJSON(Response{Type: "get_user_response", Success: false, Message: "Invalid session ID"})
				continue
			}

			// Find the session using the session ID
			session, err := ReadSession(req.SessionID)
			if err != nil {
				conn.WriteJSON(Response{Type: "get_user_response", Success: false, Message: "Session error"})
				continue
			}

			user, ok := model.FindUserByUsername(session.Username)
			if !ok {
				conn.WriteJSON(Response{Type: "get_user_response", Success: false, Message: "User not found"})
				continue
			}

			// Return user data
			conn.WriteJSON(Response{
				Type:    "get_user_response",
				Success: true,
				Data:    user,
			})

		default:
			log.Printf("Unknown message type: %s\n", pdu.Type)
			conn.WriteJSON(Response{Type: "error", Success: false, Message: "Unknown message type"})
			conn.Close() // Close the WebSocket after an unknown message type
			return
		}
	}
}
