// internal/network/server.go

package network

import (
	"encoding/json"
	"log"
	"net/http"

	"royaka/internal/player"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		var pdu Message
		if err := json.Unmarshal(msg, &pdu); err != nil {
			log.Println("PDU decode error:", err)
			continue
		}

		switch pdu.Type {
		case MsgRegister:
			var req RegisterRequest
			json.Unmarshal(pdu.Data, &req)
			err := player.AddPlayer(player.Player{Username: req.Username, Password: req.Password})
			resp := Response{Type: "register_response", Success: err == nil, Message: "Registered"}
			if err != nil {
				resp.Message = "Registration failed"
			}
			conn.WriteJSON(resp)

		case MsgLogin:
			var req LoginRequest
			json.Unmarshal(pdu.Data, &req)
			ok := player.FindPlayerByUsername(req.Username)
			resp := Response{Type: "login_response", Success: ok, Message: "Login successful"}
			if !ok {
				resp.Message = "Invalid credentials"
			}
			conn.WriteJSON(resp)
		}

	}
}
