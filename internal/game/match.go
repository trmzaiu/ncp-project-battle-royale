package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"

	"github.com/gorilla/websocket"
)

// HandleStartGame processes a request to enter matchmaking.
func HandleFindMatch(conn *websocket.Conn, data json.RawMessage) {
	var req utils.FindMatchRequest

	if err := json.Unmarshal(data, &req); err != nil || req.User == nil || req.Mode == "" {
		log.Println("[MATCH] Invalid start game request:", err)
		conn.WriteJSON(utils.Response{
			Type:    "find_match_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	username := req.User.Username
	log.Printf("[MATCH] Start request: user=%s, mode=%s", username, req.Mode)

	pendingMu.Lock()
	if pendingPlayers[username] {
		pendingMu.Unlock()
		log.Printf("[MATCH] %s already in queue", username)
		conn.WriteJSON(utils.Response{
			Type:    "find_match_response",
			Success: false,
			Message: "Already in queue",
		})
		return
	}
	pendingPlayers[username] = true
	pendingMu.Unlock()

	log.Printf("[MATCH] %s marked as pending", username)

	clientConn := &ClientConnection{Conn: conn}

	clientsMu.Lock()
	clients[username] = clientConn
	clientsMu.Unlock()

	player := model.NewPlayer(req.User, req.Mode)
	model.RegisterConnection(conn, player)

	clientConn.SafeWrite(utils.Response{
		Type:    "find_match_response",
		Success: true,
		Message: "Added to match queue. Waiting for opponent...",
	})

	log.Printf("[MATCH] %s added to queue", req.User.Username)

	go func() {
		matchQueue <- player
		log.Printf("[MATCH] %s pushed to matchQueue channel", username)

		if !matchmakerRunning {
			startMatchmaker()
			log.Println("[MATCH] Matchmaker started")
		}
	}()
}

// startMatchmaker continuously pairs players from the queue.
func startMatchmaker() {
	matchmakerRunning = true
	go func() {
		for {
			if len(matchQueue) < 2 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			player1, player2 := <-matchQueue, <-matchQueue
			log.Printf("[MATCH] Pairing %s vs %s", player1.User.Username, player2.User.Username)

			if validatePlayers(player1, player2) {
				handleMatch(player1, player2)
			}
		}
	}()
}

// validatePlayers ensures both players are still connected and unique.
func validatePlayers(p1, p2 *model.Player) bool {
	clientsMu.RLock()
	_, ok1 := clients[p1.User.Username]
	_, ok2 := clients[p2.User.Username]
	clientsMu.RUnlock()

	if !ok1 || !ok2 || p1.User.Username == p2.User.Username {
		if ok1 {
			matchQueue <- p1
		}
		if ok2 && p1.User.Username != p2.User.Username {
			matchQueue <- p2
		}
		log.Printf("[MATCH] Validation failed: %s vs %s", p1.User.Username, p2.User.Username)
		return false
	}

	return true
}

// handleMatch creates a room for the two players and notifies them.
func handleMatch(p1, p2 *model.Player) {
	clientsMu.RLock()
	conn1 := clients[p1.User.Username]
	conn2 := clients[p2.User.Username]
	clientsMu.RUnlock()

	roomID := utils.GenerateRoomID()
	room := NewRoom(roomID, p1, p2)

	roomsMu.Lock()
	rooms[roomID] = room
	roomsMu.Unlock()

	RegisterRoom(roomID, room)

	log.Printf("[ROOM] Created room %s for %s and %s", roomID, p1.User.Username, p2.User.Username)

	notifyMatchFound(conn1, p2.User.Username, roomID)
	notifyMatchFound(conn2, p1.User.Username, roomID)

	pendingMu.Lock()
	delete(pendingPlayers, p1.User.Username)
	delete(pendingPlayers, p2.User.Username)
	pendingMu.Unlock()

	log.Printf("[MATCH] %s and %s removed from pending list", p1.User.Username, p2.User.Username)
}

// notifyMatchFound tells a player that a match has been made.
func notifyMatchFound(conn *ClientConnection, opponent string, roomID string) {
	conn.SafeWrite(utils.Response{
		Type:    "match_found",
		Success: true,
		Message: "Match found!",
		Data: map[string]interface{}{
			"room_id":  roomID,
			"opponent": opponent,
		},
	})
}
