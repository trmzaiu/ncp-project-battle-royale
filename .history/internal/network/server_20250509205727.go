// internal/network/server.go

package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"royaka/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

// Session store
var store = sessions.NewCookieStore([]byte("royaka-secret"))

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ReadSession retrieves the session associated with the request.
func ReadSession(r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return nil, err
	}
	return session, nil
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
					session, err := store.Get(r, "session-name")
					if err != nil {
						log.Println("Error getting session:", err)
						conn.WriteJSON(Response{Type: "login_response", Success: false, Message: "Error creating session"})
						continue
					}

					// Store session data in the session
					session.Values["authenticated"] = true
					session.Values["username"] = req.Username
					session.Values["session_id"] = sessionID
					session.Save(r, w)

					resp.Success = true
					resp.Message = "Login successful"
					resp.Data = map[string]string{"session_id": sessionID}
					conn.WriteJSON(resp)
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
			session, err := store.Get(r, "session-name")
			if err != nil {
				conn.WriteJSON(Response{Type: "get_user_response", Success: false, Message: "Session error"})
				continue
			}

			fmt.Println("Session ID in server session:", session.Values["session_id"])
			fmt.Println("Session ID from request:", req.SessionID)

			// Get the session ID from the session store and compare
			sessionID, ok := session.Values["session_id"].(string)
			if !ok || sessionID != req.SessionID {
				conn.WriteJSON(Response{Type: "get_user_response", Success: false, Message: "Invalid session"})
				continue
			}

			// Get username from session and find user
			username, ok := session.Values["username"].(string)
			if !ok {
				conn.WriteJSON(Response{Type: "get_user_response", Success: false, Message: "Username not found in session"})
				continue
			}

			user, ok := model.FindUserByUsername(username)
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
