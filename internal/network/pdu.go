// internal/network/pdu.go

package network

import "encoding/json"

type MessageType string

const (
	MsgRegister MessageType = "register"
	MsgLogin    MessageType = "login"
)

type PDU struct {
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
	Success bool   `json:"success"`
	Message string `json:"message"`
}
