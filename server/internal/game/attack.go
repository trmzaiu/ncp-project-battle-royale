package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

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

	if room.Game.CurrentPlayer().User.Username != attacker.User.Username {
		sendToClient(attacker.User.Username, utils.Response{
			Type:    "attack_response",
			Success: false,
			Message: "It's not your turn!",
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

	success := damage > 0 || isDestroyed

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

	if defender.Towers["king"].HP <= 0 {
		winner, result := room.Game.CheckWinner()
		if result == "" {
			return
		}
		gameOverPayload := utils.Response{
			Type:    "game_over_response",
			Success: true,
			Message: result,
			Data: map[string]interface{}{
				"winner": winner,
			},
		}
		sendToClient(room.Player1.User.Username, gameOverPayload)
		sendToClient(room.Player2.User.Username, gameOverPayload)
	}
}