// internal/network/auth.go

package network

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

func handleRegister(conn *websocket.Conn, data json.RawMessage) {
	var req utils.RegisterRequest

	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("[WARN][AUTH] Invalid register data: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "register_response",
			Success: false,
			Message: "Invalid register data",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR][AUTH] Password hashing failed for %s: %v", req.Username, err)
		conn.WriteJSON(utils.Response{
			Type:    "register_response",
			Success: false,
			Message: "Error hashing password",
		})
		return
	}

	err = model.AddUser(*model.NewUser(req.Username, string(hashedPassword)))
	if err != nil {
		log.Printf("[WARN][AUTH] Registration failed for %s: %v", req.Username, err)
		conn.WriteJSON(utils.Response{
			Type:    "register_response",
			Success: false,
			Message: "Registration failed: " + err.Error(),
		})
		return
	}

	log.Printf("[INFO][AUTH] User %s registered successfully", req.Username)
	conn.WriteJSON(utils.Response{
		Type:    "register_response",
		Success: true,
		Message: "Registered successfully",
	})
}

func handleLogin(conn *websocket.Conn, data json.RawMessage) {
	var req utils.LoginRequest

	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("[WARN][AUTH] Invalid login data: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "login_response",
			Success: false,
			Message: "Invalid login data",
		})
		return
	}

	u, ok := model.FindUserByUsername(req.Username)
	if !ok {
		log.Printf("[WARN][AUTH] Login failed, user %s not found", req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "login_response",
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		log.Printf("[WARN][AUTH] Login failed, incorrect password for %s", req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "login_response",
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	sessionID := uuid.New().String()[:8]
	session := Session{SessionID: sessionID, Username: req.Username, Authenticated: true}
	log.Printf("[INFO][AUTH] User %s authenticated, session ID: %s", req.Username, sessionID)
	sessions, err := ReadSessions()
	if err != nil {
		log.Printf("[ERROR][AUTH] Reading sessions failed: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "login_response",
			Success: false,
			Message: "Error reading sessions",
		})
		return
	}

	sessions = append(sessions, session)
	if err := WriteSession(sessions); err != nil {
		log.Printf("[ERROR][AUTH] Writing sessions failed: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "login_response",
			Success: false,
			Message: "Error saving session",
		})
		return
	}

	log.Printf("[INFO][AUTH] Session stored for user %s", req.Username)
	conn.WriteJSON(utils.Response{
		Type:    "login_response",
		Success: true,
		Message: "Login successful",
		Data:    map[string]string{"session_id": sessionID},
	})
}

func handleGetUser(conn *websocket.Conn, data json.RawMessage) {
	var req utils.UserRequest

	if err := json.Unmarshal(data, &req); err != nil {
		log.Printf("[WARN][AUTH] Invalid session ID in get_user: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "user_response",
			Success: false,
			Message: "Invalid session ID",
		})
		return
	}

	session, err := FindSessionByID(req.SessionID)
	if err != nil {
		log.Printf("[WARN][AUTH] Session %s not found", req.SessionID)
		conn.WriteJSON(utils.Response{
			Type:    "user_response",
			Success: false,
			Message: "Session not found",
		})
		return
	}

	user, ok := model.FindUserByUsername(session.Username)
	if !ok {
		log.Printf("[WARN][AUTH] User %s from session not found", session.Username)
		conn.WriteJSON(utils.Response{
			Type:    "user_response",
			Success: false,
			Message: "User not found",
		})
		return
	}

	log.Printf("[INFO][AUTH] Returning user data for %s (session %s)", user.Username, req.SessionID)
	conn.WriteJSON(utils.Response{
		Type:    "user_response",
		Success: true,
		Data: map[string]interface{}{
			"user":   user,
			"maxExp": model.GetMaxExp(user.Level),
		},
	})
}
