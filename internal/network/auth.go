// internal/network/auth.go

package network

import (
	"encoding/json"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

func handleRegister(conn *websocket.Conn, data json.RawMessage) {
	var req utils.RegisterRequest
	if err := json.Unmarshal(data, &req); err != nil {
		conn.WriteJSON(utils.Response{Type: "register_response", Success: false, Message: "Invalid register data"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		conn.WriteJSON(utils.Response{Type: "register_response", Success: false, Message: "Error hashing password"})
		return
	}

	err = model.AddUser(*model.NewUser(req.Username, string(hashedPassword)))
	resp := utils.Response{Type: "register_response", Success: err == nil, Message: "Registered successfully"}
	if err != nil {
		resp.Message = "Registration failed: " + err.Error()
	}
	conn.WriteJSON(resp)
}

func handleLogin(conn *websocket.Conn, data json.RawMessage) {
	var req utils.LoginRequest
	if err := json.Unmarshal(data, &req); err != nil {
		conn.WriteJSON(utils.Response{Type: "login_response", Success: false, Message: "Invalid login data"})
		return
	}

	u, ok := model.FindUserByUsername(req.Username)
	if !ok || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		conn.WriteJSON(utils.Response{Type: "login_response", Success: false, Message: "Invalid credentials"})
		return
	}

	sessionID := uuid.New().String()[:8]
	session := Session{SessionID: sessionID, Username: req.Username, Authenticated: true}

	sessions, err := ReadSessions()
	if err != nil {
		conn.WriteJSON(utils.Response{Type: "login_response", Success: false, Message: "Error reading sessions"})
		return
	}

	sessions = append(sessions, session)
	if err := WriteSession(sessions); err != nil {
		conn.WriteJSON(utils.Response{Type: "login_response", Success: false, Message: "Error saving session"})
		return
	}

	conn.WriteJSON(utils.Response{
		Type:    "login_response",
		Success: true,
		Message: "Login successful",
		Data:    map[string]string{"session_id": sessionID},
	})
}

func handleGetUser(conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	
	if err := json.Unmarshal(data, &req); err != nil {
		conn.WriteJSON(utils.Response{Type: "get_user_response", Success: false, Message: "Invalid session ID"})
		return
	}

	session, err := FindSessionByID(req.SessionID)
	if err != nil {
		conn.WriteJSON(utils.Response{Type: "get_user_response", Success: false, Message: "Session not found"})
		return
	}

	user, ok := model.FindUserByUsername(session.Username)
	if !ok {
		conn.WriteJSON(utils.Response{Type: "get_user_response", Success: false, Message: "User not found"})
		return
	}

	conn.WriteJSON(utils.Response{
		Type:    "get_user_response",
		Success: true,
		Data:    user,
	})
}
