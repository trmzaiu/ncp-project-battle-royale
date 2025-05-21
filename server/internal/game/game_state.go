package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandleGetGame(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameRequest

	// Parse & validate request
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" {
		log.Printf("[WARN][GAME] invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	// Get room safely
	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !exists {
		log.Printf("[WARN][GAME] room %s not found for user %s", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	// Identify current player and opponent
	var currentUser, opponent *model.Player
	if room.Player1.User.Username == req.Username {
		currentUser, opponent = room.Player1, room.Player2
	} else if room.Player2.User.Username == req.Username {
		currentUser, opponent = room.Player2, room.Player1
	} else {
		log.Printf("[WARN][GAME] user %s not in room %s", req.Username, req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: "Player not in room",
		})
		return
	}

	conn.WriteJSON(utils.Response{
		Type:    "game_response",
		Success: true,
		Message: "Game info loaded",
		Data: map[string]interface{}{
			"user":     currentUser,
			"opponent": opponent,
			"turn":     room.Game.Turn,
			"player1":  room.Player1.User.Username,
			"map":      room.Game.BattleMap,
		},
	})

	log.Printf("[INFO][GAME] sent game state to %s in room %s", req.Username, req.RoomID)
}

func HandleGetGameEnhanced(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameRequest

	// Parse & validate request
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" {
		log.Printf("[WARN][GAME] invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	// Get room safely
	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !exists {
		log.Printf("[WARN][GAME] room %s not found for user %s", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	// Identify current player and opponent
	var currentUser, opponent *model.Player
	if room.Player1.User.Username == req.Username {
		currentUser, opponent = room.Player1, room.Player2
	} else if room.Player2.User.Username == req.Username {
		currentUser, opponent = room.Player2, room.Player1
	} else {
		log.Printf("[WARN][GAME] user %s not in room %s", req.Username, req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: "Player not in room",
		})
		return
	}

	conn.WriteJSON(utils.Response{
		Type:    "game_response",
		Success: true,
		Message: "Game info loaded",
		Data: map[string]interface{}{
			"user":     currentUser,
			"opponent": opponent,
		},
	})

	log.Printf("[INFO][GAME] sent game state to %s in room %s", req.Username, req.RoomID)
}
