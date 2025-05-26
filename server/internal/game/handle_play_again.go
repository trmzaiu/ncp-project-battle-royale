package game

import (
	"encoding/json"
	"log"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandlePlayAgain(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameOverRequest

	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" {
		log.Printf("[WARN][PLAY_AGAIN] Invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "play_again_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !exists {
		log.Printf("[WARN][PLAY_AGAIN] Room %s not found", req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "play_again_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	roomsMu.Lock()
	delete(rooms, room.ID)
	roomsMu.Unlock()

	log.Printf("[INFO][PLAY_AGAIN] Room %s cleaned up", room.ID)
}