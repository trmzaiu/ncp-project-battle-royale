// internal/network/server.go

package network

import (
	"encoding/json"
	"log"
	"net/http"

	"royaka/internal/game"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWS handles WebSocket connections and messages
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in WebSocket handler:", r)
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			return
		}

		var pdu utils.Message
		if err := json.Unmarshal(msg, &pdu); err != nil {
			conn.WriteJSON(utils.Response{Type: "error", Success: false, Message: "Invalid message format"})
			continue
		}

		switch pdu.Type {
		case "register":
			handleRegister(conn, pdu.Data)
		case "login":
			handleLogin(conn, pdu.Data)
		case "get_user":
			handleGetUser(conn, pdu.Data)
		case "start_game":
			game.HandleStartGame(conn, pdu.Data)
		case "get_game_info":
			game.HandleGetGameInfo(conn, pdu.Data)
		default:
			conn.WriteJSON(utils.Response{Type: "error", Success: false, Message: "Unknown message type"})
			conn.Close()
			return
		}
	}
}
