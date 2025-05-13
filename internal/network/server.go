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
	log.Println("[WS] New WebSocket connection request")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("[ATTACK] Error closing connection: %v", err)
		}
		log.Printf("[ATTACK] WebSocket connection for client is closed")
	}()	

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WS] Recovered from panic: %v", r)
		}
	}()

	log.Println("[WS] WebSocket connection established")

	for {
		if !readAndProcessMessage(conn) {
			break
		}
	}
}

func readAndProcessMessage(conn *websocket.Conn) bool {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		logWebSocketError(err)
		return false
	}

	log.Printf("[WS] Received raw message: %s", msg)

	var pdu utils.Message
	if err := json.Unmarshal(msg, &pdu); err != nil {
		log.Printf("[WS] JSON unmarshal error: %v", err)
		sendError(conn, "Invalid message format")
		return true
	}

	log.Printf("[WS] Handling message type: %s", pdu.Type)
	processMessage(conn, pdu)
	return true
}

func processMessage(conn *websocket.Conn, pdu utils.Message) {
	switch pdu.Type {
	case "register":
		handleRegister(conn, pdu.Data)
	case "login":
		handleLogin(conn, pdu.Data)
	case "get_user":
		handleGetUser(conn, pdu.Data)
	case "find_match":
		game.HandleFindMatch(conn, pdu.Data)
	case "get_game":
		game.HandleGetGame(conn, pdu.Data)
	case "attack":
		game.HandleAttack(conn, pdu.Data)
	default:
		log.Printf("[WS] Unknown message type: %s", pdu.Type)
		sendError(conn, "Unknown message type")
	}
}

func sendError(conn *websocket.Conn, message string) {
    if err := conn.WriteJSON(utils.Response{
        Type:    "error",
        Success: false,
        Message: message,
    }); err != nil {
        log.Printf("[WS] Error sending error message: %v", err)
    }
}

func logWebSocketError(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("[WS] Unexpected WebSocket closure: %v", err)
	} else {
		log.Printf("[WS] Error reading message: %v", err)
	}
}
