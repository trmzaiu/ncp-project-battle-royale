package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandleSelectTroop(conn *websocket.Conn, data json.RawMessage) {
	var req utils.SelectTroopRequest

	// Parse & validate request
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" || req.Troop == "" {
		log.Printf("[WARN][SELECT] Invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !exists {
		log.Printf("[WARN][SELECT] Room %s not found", req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	var player *model.Player
	if room.Player1.User.Username == req.Username {
		player = room.Player1
	} else if room.Player2.User.Username == req.Username {
		player = room.Player2
	} else {
		log.Printf("[WARN][SELECT] User %s not found in room", req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "You are not part of this match",
		})
		return
	}

	// Validate troop is in player's hand
	found := false
	for _, t := range player.Troops {
		if t.Name == req.Troop {
			found = true
			break
		}
	}
	if !found {
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "Troop not in current hand",
		})
		return
	}

	// Rotate the troop
	player.RotateTroop(req.Troop)

	// Send update to both players
	payload := utils.Response{
		Type:    "troop_response",
		Success: true,
		Message: "Troop selected and rotated",
		Data: map[string]interface{}{
			"player": player,
		},
	}

	sendToClient(room.Player1.User.Username, payload)
	sendToClient(room.Player2.User.Username, payload)
}