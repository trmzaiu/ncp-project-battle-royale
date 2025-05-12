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

	log.Printf("[GAME_INFO] Received request: %s", string(data))
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" {
		log.Printf("[GAME_INFO] Invalid request payload")
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	roomsMu.Lock()
	room, exists := rooms[req.RoomID]
	roomsMu.Unlock()
	if !exists {
		log.Printf("[GAME_INFO] Room %s not found for user %s", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: "Room not found",
		})
		return
	}

	var currentUser, opponent *model.Player
	if room.Player1.User.Username == req.Username {
		currentUser, opponent = room.Player1, room.Player2
	} else if room.Player2.User.Username == req.Username {
		currentUser, opponent = room.Player2, room.Player1
	}

	if currentUser == nil || opponent == nil {
		log.Printf("[GAME_INFO] Player %s not part of room %s", req.Username, req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "game_response",
			Success: false,
			Message: "Invalid player in room",
		})
		return
	}

	log.Printf("[GAME_INFO] Sending game state to %s in room %s", req.Username, req.RoomID)

	conn.WriteJSON(utils.Response{
		Type:    "game_response",
		Success: true,
		Message: "Game info loaded",
		Data: map[string]interface{}{
			"user":      currentUser,
			"opponent":  opponent,
			"your_turn": (room.Game.Turn == 1 && currentUser == room.Player1) || (room.Game.Turn == 2 && currentUser == room.Player2),
		},
	})
}

// HandleAttack processes a player's attack action.
func HandleAttack(conn *websocket.Conn, data json.RawMessage) {
	var req utils.AttackRequest

	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" || req.Troop == "" || req.Target == "" {
		log.Printf("[ATTACK] Invalid attack request")
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	roomsMu.Lock()
	room, exists := rooms[req.RoomID]
	roomsMu.Unlock()

	if !exists {
		log.Printf("[ROOM] Room %s not found", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: "Room not found",
		})
		return
	}

	var attacker *model.Player
	if room.Player1.User.Username == req.Username {
		attacker = room.Player1
	} else if room.Player2.User.Username == req.Username {
		attacker = room.Player2
	} else {
		return
	}

	var troop *model.Troop
	for i := range attacker.Troops {
		if attacker.Troops[i].Name == req.Troop {
			troop = attacker.Troops[i]
			break
		}
	}
	if troop == nil {
		return
	}

	log.Printf("[ATTACK] %s attacking %s using troop %s in room %s", req.Username, req.Target, req.Troop, req.RoomID)
	result := room.Game.PlayTurn(attacker, troop, req.Target)
	if result == "" {
		log.Printf("[ATTACK] Invalid attack result", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: "Invalid attack result",
		})
		return
	}

	log.Printf("[ATTACK] Result: %s -> %s", req.Username, result)

	payload := utils.Response{
		Type:    "attack_response",
		Success: true,
		Message: result,
		Data: map[string]interface{}{
			"attacker": req.Username,
			"troop":    req.Troop,
			"target":   req.Target,
			"result":   result,
		},
	}

	log.Printf(room.Player1.User.Username, room.Player2.User.Username)
	conn.WriteJSON(payload)

	clientsMu.Lock()
	client1 := clients[room.Player1.User.Username]
	client2 := clients[room.Player2.User.Username]
	clientsMu.Unlock()

	sendToClient(client1, payload)
	sendToClient(client2, payload)
}

func getClientsForRoom(room *Room) (*ClientConnection, *ClientConnection) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	client1 := clients[room.Player1.User.Username]
	client2 := clients[room.Player2.User.Username]
	return client1, client2
}

func sendToClient(client *ClientConnection, payload utils.Response) {
	if client != nil && client.Conn != nil {
		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
            log.Printf("[ATTACK] WebSocket ping failed: %v", err)
            return
        }
		client.SafeWrite(payload)
	} else {
		log.Printf("[ATTACK] Connection for client is closed")
	}
}

func NotifyGameConclusion(room *Room, winner *model.Player) {
	log.Printf("[GAME_END] Winner is %s in room %s", winner.User.Username, room.ID)
	message := utils.Response{
		Type:    "game_finished",
		Success: true,
		Message: "Game over! " + winner.User.Username + " wins!",
	}

	// Manage clients using channel-based synchronization
	clientsMu.Lock()
	client1 := clients[room.Player1.User.Username]
	client2 := clients[room.Player2.User.Username]
	clientsMu.Unlock()

	if client1 != nil {
		client1.SafeWrite(message)
	}
	if client2 != nil {
		client2.SafeWrite(message)
	}
}
