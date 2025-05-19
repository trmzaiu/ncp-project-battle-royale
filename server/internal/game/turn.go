package game

import (
	"encoding/json"
	"log"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandleSkipTurn(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameRequest
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" {
		log.Printf("[WARN][SKIP_TURN] Invalid request from conn %v: %v | Data: %s", conn.RemoteAddr(), err, string(data))
		conn.WriteJSON(utils.Response{
			Type:    "skip_turn_response",
			Success: false,
			Message: "Invalid skip turn request",
		})
		return
	}

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()

	if !exists {
		log.Printf("[WARN][SKIP_TURN] Room not found: %s by user %s", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "skip_turn_response",
			Success: false,
			Message: "Room not found",
		})
		return
	}

	current := room.Game.CurrentPlayer()
	if current.User.Username != req.Username {
		log.Printf("[WARN][SKIP_TURN] Not %s's turn in room %s", req.Username, req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "skip_turn_response",
			Success: false,
			Message: "It's not your turn!",
		})
		return
	}

	room.Game.SkipTurn(current)

	log.Printf("[DEBUG][SKIP_TURN] Turn switched to: %s", room.Game.Turn)

	payload := utils.Response{
		Type:    "skip_turn_response",
		Success: true,
		Message: "Turn skipped",
		Data: map[string]interface{}{
			"turn":    room.Game.Turn,
			"player1": room.Game.Player1,
			"player2": room.Game.Player2,
		},
	}

	sendToClient(room.Player1.User.Username, payload)
	sendToClient(room.Player2.User.Username, payload)
}