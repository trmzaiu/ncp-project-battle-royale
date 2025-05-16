// internal/utils/pdu.go

package utils

import (
	"encoding/json"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Response struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRequest struct {
	SessionID string `json:"session_id"`
}

type FindMatchRequest struct {
	Username string `json:"username"`
	Mode     string `json:"mode"`
}

type GameRequest struct {
	RoomID   string `json:"room_id"`
	Username string `json:"username"`
}

type AttackRequest struct {
	RoomID   string `json:"room_id"`
	Username string `json:"username"`
	Troop    string `json:"troop"`
	Target   string `json:"target"`
}

type GameOverRequest struct {
	RoomID string `json:"room_id"`
}
