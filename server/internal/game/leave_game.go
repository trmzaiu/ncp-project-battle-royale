package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandleLeaveGame(conn *websocket.Conn, data json.RawMessage) {
	var req utils.GameRequest

	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" {
		conn.WriteJSON(utils.Response{
			Type:    "leave_game_response",
			Success: false,
			Message: "",
		})
		return
	}

	room, found := rooms[req.RoomID]
	if !found {
		conn.WriteJSON(utils.Response{
			Type:    "leave_game_response",
			Success: false,
			Message: "Room not found",
		})
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	player1 := room.Game.Player1
	player2 := room.Game.Player2

	var winner *model.Player

	if player1 != nil && player1.User.Username == req.Username {
		winner = player2
	} else if player2 != nil && player2.User.Username == req.Username {
		winner = player1
	}

	if winner != nil {
		room.Game.SetWinner(winner)

		payload := utils.Response{
			Type:    "game_over_response",
			Success: true,
			Message: "",
			Data: map[string]interface{}{
				"winner": winner,
			},
		}

		sendToClient(winner.User.Username, payload)
	}

	conn.WriteJSON(utils.Response{
		Type:    "leave_game_response",
		Success: true,
		Message: "Left room and winner set if applicable",
	})
	room.Game.StopGameLoop()
}

func HandleDisconnect(conn *websocket.Conn) {
	player := model.GetPlayerByConn(conn)
	if player == nil {
		return
	}

	username := player.User.Username
	roomID := GetRoomIDByUsername(username)
	if roomID == "" {
		return
	}

	log.Printf("[INFO] %s disconnected, handling leave...", username)

	req := utils.GameRequest{
		RoomID:   roomID,
		Username: username,
	}
	raw, _ := json.Marshal(req)
	HandleLeaveGame(conn, raw)
}
