// internal/game/match.go

package game

import (
	"encoding/json"
	"fmt"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[string]*websocket.Conn)
	clientsMu sync.Mutex

	matchQueue = make(chan *model.Player, 100)
	rooms      = make(map[string]*model.Room)
	matchmakerRunning bool
)

// handleStartGame handles player joining the match queue
func HandleStartGame(conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		Username string `json:"username"`
	}

	if err := json.Unmarshal(data, &req); err != nil || req.Username == "" {
		conn.WriteJSON(utils.Response{Type: "start_game_response", Success: false, Message: "Invalid username"})
		return
	}

	player := model.NewPlayer(req.Username)

	clientsMu.Lock()
	clients[req.Username] = conn
	clientsMu.Unlock()

	go func() {
		matchQueue <- player
		log.Printf("Player %s added to match queue", req.Username)

		if len(matchQueue) >= 2 && !matchmakerRunning {
			log.Println("Starting matchmaker...")
			startMatchmaker()
		}
	}()

	conn.WriteJSON(utils.Response{
		Type:    "start_game_response",
		Success: true,
		Message: "Player created and added to match queue. Waiting for opponent...",
	})
}

func startMatchmaker() {
	matchmakerRunning = true
	go func() {
		for {
			player1 := <-matchQueue
			player2 := <-matchQueue
			startMatch(player1, player2)
		}
	}()
}

func startMatch(p1, p2 *model.Player) {
	log.Printf("Match started between %s and %s", p1.Username, p2.Username)

	roomID := generateRoomID()
	room := model.NewRoom(roomID, p1, p2)
	rooms[roomID] = room

	clientsMu.Lock()
	conn1 := clients[p1.Username]
	conn2 := clients[p2.Username]
	clientsMu.Unlock()

	matchInfo := map[string]interface{}{
		"room_id":  roomID,
		"opponent": p2.Username,
	}
	conn1.WriteJSON(utils.Response{
		Type:    "match_found",
		Success: true,
		Message: "Match found!",
		Data:    matchInfo,
	})

	matchInfo["opponent"] = p1.Username
	conn2.WriteJSON(utils.Response{
		Type:    "match_found",
		Success: true,
		Message: "Match found!",
		Data:    matchInfo,
	})

	NewGame(p1, p2, "simple")
}

func generateRoomID() string {
	return fmt.Sprintf("room-%d", time.Now().UnixNano())
}
