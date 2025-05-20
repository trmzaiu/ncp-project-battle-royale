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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR][WS] Upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}

	// Recover panic inside the goroutine safely
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR][WS] Panic recovered: %v", r)
		}
		if err := conn.Close(); err != nil {
			log.Printf("[ERROR][WS] Connection close failed: %v", err)
		}
		// game.HandleDisconnect(conn)
		log.Println("[WS] Connection closed")
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

	log.Printf("[INFO][WS] Raw message: %s", msg)

	var pdu utils.Message
	if err := json.Unmarshal(msg, &pdu); err != nil {
		log.Printf("[WARN][WS] Invalid JSON: %v", err)
		sendError(conn, "Invalid message format")
		return true
	}

	log.Printf("[INFO][WS] Message type: %s", pdu.Type)
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
	case "heal":
		game.HandleHeal(conn, pdu.Data)
	case "skip_turn":
		game.HandleSkipTurn(conn, pdu.Data)
	case "play_again":
		game.HandlePlayAgain(conn, pdu.Data)
	// case "leave_game":
	// 	game.HandleLeaveGame(conn, pdu.Data)
	case "select_troop":
		game.HandleSelectTroop(conn, pdu.Data)
	default:
		log.Printf("[WARN][WS] Unknown message type: %s", pdu.Type)
		sendError(conn, "Unknown message type")
	}
}

func sendError(conn *websocket.Conn, message string) {
	err := conn.WriteJSON(utils.Response{
		Type:    "error",
		Success: false,
		Message: message,
	})
	if err != nil {
		log.Printf("[ERROR][WS] Failed to send error response: %v", err)
	}
}

func logWebSocketError(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("[ERROR][WS] Unexpected closure: %v", err)
	} else {
		log.Printf("[WARN][WS] Client disconnected: %v", err)
	}
}
