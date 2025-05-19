package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandleHeal(conn *websocket.Conn, data json.RawMessage) {
	var req utils.HealRequest

	// Parse & validate request data
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" || req.Troop == "" {
		log.Printf("[WARN][HEAL] invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "heal_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	// Fetch the room from memory
	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !exists {
		log.Printf("[WARN][HEAL] Room %s not found for user %s", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "heal_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	// Identify the player
	var player, opponent *model.Player
	if room.Player1.User.Username == req.Username {
		player = room.Player1
		opponent = room.Player2
	} else if room.Player2.User.Username == req.Username {
		player = room.Player2
		opponent = room.Player1
	} else {
		log.Printf("[WARN][HEAL] User %s not in room %s", req.Username, req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "heal_response",
			Success: false,
			Message: "You are not part of this match",
		})
		return
	}

	// Find the troop being used for healing
	var troop *model.Troop
	for i := range player.Troops {
		if player.Troops[i].Name == req.Troop {
			troop = player.Troops[i]
			break
		}
	}
	if troop == nil {
		log.Printf("[WARN][HEAL] Troop %s not found for user %s", req.Troop, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "heal_response",
			Success: false,
			Message: "Invalid troop used for healing",
		})
		return
	}

	// Call the heal method
	actualHealed, healedTower, message := room.Game.HealTower(player, troop)
	if actualHealed == 0 {
		conn.WriteJSON(utils.Response{
			Type:    "heal_response",
			Success: false,
			Message: message,
		})
		return
	}

	// Prepare response payload
	payload := utils.Response{
		Type:    "heal_response",
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"player":      player,
			"opponent":    opponent,
			"troop":       troop.Name,
			"healedTower": healedTower,
			"healAmount":  actualHealed,
			"turn":        room.Game.Turn,
		},
	}

	// Broadcast to both players
	sendToClient(room.Player1.User.Username, payload)
	sendToClient(room.Player2.User.Username, payload)
}