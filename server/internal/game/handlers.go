package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"

	"github.com/gorilla/websocket"
)

// HandleGetGame sends the current game state to the requesting player.
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
		},
	})

	log.Printf("[INFO][GAME] sent game state to %s in room %s", req.Username, req.RoomID)
}

// HandleAttack processes a player's attack action.
func HandleAttack(conn *websocket.Conn, data json.RawMessage) {
	var req utils.AttackRequest

	// Parse & validate request data
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" || req.Troop == "" || req.Target == "" {
		log.Printf("[WARN][ATTACK] invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
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
		log.Printf("[WARN][ATTACK] Room %s not found for user %s", req.RoomID, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	// Identify the attacker
	var attacker, defender *model.Player
	if room.Player1.User.Username == req.Username {
		attacker = room.Player1
		defender = room.Player2
	} else if room.Player2.User.Username == req.Username {
		attacker = room.Player2
		defender = room.Player1
	} else {
		log.Printf("[WARN][ATTACK] User %s not in room %s", req.Username, req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: "You are not part of this match",
		})
		return
	}

	// Find the troop being used for attack
	var troop *model.Troop
	for i := range attacker.Troops {
		if attacker.Troops[i].Name == req.Troop {
			troop = attacker.Troops[i]
			break
		}
	}
	if troop == nil {
		log.Printf("[WARN][ATTACK] Troop %s not found for user %s", req.Troop, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: "Invalid troop used for attack",
		})
		return
	}

	// Process the attack via game logic
	log.Printf("[INFO][ATTACK] %s attacking with %s targeting %s in room %s", attacker.User.Username, troop.Name, req.Target, req.RoomID)
	damage, isCrit, message := room.Game.PlayTurnSimple(attacker, troop, req.Target)
	isDestroyed := defender.Towers[req.Target].HP <= 0

	success := damage > 0 || isCrit || isDestroyed

	payload := utils.Response{
		Type:    "attack_response",
		Success: success,
		Message: message,
		Data: map[string]interface{}{
			"attacker":    attacker,
			"defender":    defender,
			"troop":       troop.Name,
			"target":      req.Target,
			"damage":      damage,
			"isCrit":      isCrit,
			"isDestroyed": isDestroyed,
			"turn":        room.Game.Turn,
		},
	}

	sendToClient(room.Player1.User.Username, payload)
	sendToClient(room.Player2.User.Username, payload)
}

func sendToClient(username string, payload utils.Response) {
	clientsMu.RLock()
	client, exists := clients[username]
	clientsMu.RUnlock()

	if !exists || client == nil || client.Conn == nil {
		log.Printf("[WARN][SEND] Client %s not found or connection is nil", username)
		return
	}

	go func() {
		if err := client.SafeWrite(payload); err != nil {
			log.Printf("[ERROR][SEND] Failed to send to %s: %v", username, err)
		}
	}()
}

func NotifyGameConclusion(room *Room, winner *model.Player) {
	log.Printf("[INFO][GAME_OVER] Winner: %s in room %s", winner.User.Username, room.ID)
	message := utils.Response{
		Type:    "game_over_response",
		Success: true,
		Message: "Game over! " + winner.User.Username + " wins!",
		Data: map[string]interface{}{
			"winner": winner.User.Username,
		},
	}
	sendToClient(room.Player1.User.Username, message)
	sendToClient(room.Player2.User.Username, message)
}

func HandleSkipTurn(conn *websocket.Conn, data json.RawMessage) {
	var req utils.SkipTurnRequest
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

func HandleGameOver(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameOverRequest

	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" {
		log.Printf("[WARN][GAME_OVER] Invalid request: %v", err)
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
		log.Printf("[WARN][GAME_OVER] Room %s not found", req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "game_over_response",
			Success: false,
			Message: roomRequestMessage,
		})
		return
	}

	if result := room.Game.CheckWinner(); result == "" {
		log.Printf("[INFO][GAME_OVER] Game still in progress in room %s", req.RoomID)
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
		if p1.DestroyedCount() > p2.DestroyedCount() {
			winner = p1
		} else if p2.DestroyedCount() > p1.DestroyedCount() {
			winner = p2
		}
	}

	if winner != nil {
		NotifyGameConclusion(room, winner)
	} else {
		msg := utils.Response{
			Type:    "game_over_response",
			Success: true,
			Message: "Game over! It's a draw!",
		}
		sendToClient(p1.User.Username, msg)
		sendToClient(p2.User.Username, msg)
		log.Printf("[INFO][GAME_OVER] Room %s ended in draw", room.ID)
	}
}

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
