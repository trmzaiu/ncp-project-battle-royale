package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"

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

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
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
			"your_turn": room.Game.Turn,
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

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()

	if !exists {
		log.Printf("[ROOM] Room %s not found", req.RoomID)
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
	result, dmg := room.Game.PlayTurn(attacker, troop, req.Target)
	if result == "" {
		log.Printf("[ATTACK] Invalid attack result")
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: "Invalid attack result",
		})
		return
	}

	log.Printf("[ATTACK] Result: %s -> %s", req.Username, result)

	remainHp := attacker.Towers[req.Target].HP - dmg
	if remainHp <= 0 {
		attacker.Towers[req.Target].HP = 0
	}

	payload := utils.Response{
		Type:    "attack_response",
		Success: true,
		Message: result,
		Data: map[string]interface{}{
			"attacker":    attacker.User.Username,
			"troop":       troop.Name,
			"target":      req.Target,
			"result":      result,
			"damage":      dmg,
			"remainHp":    remainHp,
			"isDestroyed": attacker.Towers[req.Target].HP <= 0,
			"turn":        room.Game.Turn,
		},
	}

	log.Printf("[ATTACK] The remaining Hp of the tower is %d - %d = %d", attacker.Towers[req.Target].MaxHP, dmg, remainHp)

	sendToClient(room.Player1.User.Username, payload)
	sendToClient(room.Player2.User.Username, payload)
}

func sendToClient(username string, payload utils.Response) {
	clientsMu.RLock()
	client, exists := clients[username]
	clientsMu.RUnlock()
	log.Printf("[ATTACK] Sending message to clients: %v", clients)

	if !exists || client == nil || client.Conn == nil {
		log.Printf("[ATTACK] Client %s not found or connection is nil", username)
		return
	}

	go func() {
		if err := client.SafeWrite(payload); err != nil {
			log.Printf("[ATTACK] Failed to send message to client %s: %v", username, err)
		}
	}()
}

func NotifyGameConclusion(room *Room, winner *model.Player) {
	log.Printf("[GAME_END] Winner is %s in room %s", winner.User.Username, room.ID)
	message := utils.Response{
		Type:    "game_finished",
		Success: true,
		Message: "Game over! " + winner.User.Username + " wins!",
	}

	// Manage clients using channel-based synchronization
	clientsMu.RLock()
	client1 := clients[room.Player1.User.Username]
	client2 := clients[room.Player2.User.Username]
	clientsMu.RUnlock()

	if client1 != nil {
		client1.SafeWrite(message)
	}
	if client2 != nil {
		client2.SafeWrite(message)
	}
}

func HandleGameOver(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameOverRequest

	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" {
		log.Printf("[GAME_OVER] Invalid request data: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "game_over_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !exists {
		log.Printf("[GAME_OVER] Room %s not found", req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "game_over_response",
			Success: false,
			Message: "Room not found",
		})
		return
	}

	if result := room.Game.CheckWinner(); result == "" {
		log.Println("[GAME_OVER] Game not finished yet for RoomID:", req.RoomID)
		return
	}

	var winner *model.Player
	p1, p2 := room.Game.Player1, room.Game.Player2

	switch {
	case p1.Towers["king"].HP <= 0:
		winner = p2
	case p2.Towers["king"].HP <= 0:
		winner = p1
	case room.Game.Enhanced && time.Since(room.Game.StartTime) > room.Game.MaxTime:
		p1Score, p2Score := p1.DestroyedCount(), p2.DestroyedCount()
		if p1Score > p2Score {
			winner = p1
		} else if p2Score > p1Score {
			winner = p2
		}
	}

	if winner != nil {
		NotifyGameConclusion(room, winner)
	} else {
		// It’s a draw
		msg := utils.Response{
			Type:    "game_over_response",
			Success: true,
			Message: "Game over! It's a draw!",
		}
		sendToClient(p1.User.Username, msg)
		sendToClient(p2.User.Username, msg)
	}

	// Clean up the room
	roomsMu.Lock()
	delete(rooms, room.ID)
	roomsMu.Unlock()

	log.Printf("[GAME_OVER] Room %s closed", room.ID)
}
