package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func HandleSelectTroop(conn *websocket.Conn, data json.RawMessage) {
	var req utils.SelectTroopRequest

	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" || req.Troop == "" {
		log.Printf("[ERROR][SELECT] Invalid request: %+v", req)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	log.Printf("[INFO][SELECT] %s is trying to spawn %s at (%f, %f) in room %s",
		req.Username, req.Troop, req.X, req.Y, req.RoomID)

	roomsMu.RLock()
	room, ok := rooms[req.RoomID]
	roomsMu.RUnlock()
	if !ok {
		log.Printf("[WARN][SELECT] Room %s not found", req.RoomID)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "Room not found",
		})
		return
	}

	var player *model.Player
	if room.Player1.User.Username == req.Username {
		player = room.Player1
	} else if room.Player2.User.Username == req.Username {
		player = room.Player2
	} else {
		log.Printf("[WARN][SELECT] %s is not in the match", req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "You are not in this match",
		})
		return
	}

	var selectedTemplate *model.Troop
	for i, t := range player.Troops {
		if t.Name == req.Troop {
			selectedTemplate = player.Troops[i]
			break
		}
	}
	if selectedTemplate == nil {
		log.Printf("[WARN][SELECT] Troop %s not found in %s's hand", req.Troop, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "Troop not in hand",
		})
		return
	}

	realX, realY := float64(req.X), float64(req.Y)
	if room.Player1.User.Username == req.Username {
		realX = 20.0 - req.X
		realY = 20.0 - req.Y
	}

	log.Printf("[INFO][SPAWN] %s spawned %s at (%f, %f)", req.Username, selectedTemplate.Name, realX, realY)

	if !room.Game.IsValidSpawnPosition(req.Username, realX, realY) {
		log.Printf("[WARN][SELECT] Invalid position (%f, %f) for %s", realX, realY, req.Username)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "Invalid spawn position",
		})
		return
	}

	// Check mana
	if room.Game.Enhanced && player.Mana < selectedTemplate.MANA {
		log.Printf("[WARN][SELECT] Not enough mana for %s to use %s (has %d, needs %d)",
			req.Username, selectedTemplate.Name, player.Mana, selectedTemplate.MANA)
		conn.WriteJSON(utils.Response{
			Type:    "troop_response",
			Success: false,
			Message: "Not enough mana",
		})
		return
	}

	player.RotateTroop(req.Troop)

	// Tạo troop instance
	instance := model.TroopInstance{
		ID:             uuid.New().String(),
		Template:       selectedTemplate,
		TypeEntity:     "troop",
		Owner:          player.User.Username,
		Position:       model.Position{X: realX, Y: realY},
		IsDead:         false,
		LastAttackTime: time.Now(),
	}

	room.Game.BattleMap = append(room.Game.BattleMap, &instance)

	// Gửi lại troop mới spawn cho cả 2
	payload := utils.Response{
		Type:    "troop_response",
		Success: true,
		Message: "Troop spawned",
		Data: map[string]interface{}{
			"troop":  instance,
			"player": player,
			"map":    room.Game.BattleMap,
		},
	}
	sendToClient(room.Player1.User.Username, payload)
	sendToClient(room.Player2.User.Username, payload)
}

func (g *Game) IsValidSpawnPosition(username string, x, y float64) bool {
	if x < 0 || x >= 21 || y < 0 || y >= 21 {
		log.Printf("[INVALID_POS] (%f, %f) is out of bounds", x, y)
		return false
	}

	for _, troop := range g.BattleMap {
		pos := troop.GetPosition()
		if pos.X == x && pos.Y == y {
			log.Printf("[INVALID_POS] (%f, %f) already occupied", x, y)
			return false
		}
	}

	if g.Player1.User.Username == username && y > 8 {
		return false
	}
	if g.Player2.User.Username == username && y < 12 {
		return false
	}

	return true
}
