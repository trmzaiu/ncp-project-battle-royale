// internal/network/pdu.go

package network

import "encoding/json"

type MessageType string

const (
	MsgRegister MessageType = "register"
	MsgLogin    MessageType = "login"
)

type Message struct {
	Type MessageType   `json:"type"`
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
    Type    string `json:"type"`    // add this
    Success bool   `json:"success"`
    Message string `json:"message"`
    Token string `json:"token,omitempty"`
}