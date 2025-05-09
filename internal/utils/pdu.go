// internal/utils/pdu.go

package utils

import "encoding/json"

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Type    string 		`json:"type"`
	Success bool   		`json:"success"`
	Message string 		`json:"message"`
	Data    any			`json:"data,omitempty"`
}