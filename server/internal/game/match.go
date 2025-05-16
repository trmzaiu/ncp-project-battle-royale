package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"

	"github.com/gorilla/websocket"
)

func HandleFindMatch(conn *websocket.Conn, data json.RawMessage) {
	var req utils.FindMatchRequest

	// Parse & validate request data
	if err := json.Unmarshal(data, &req); err != nil || req.Username == "" || req.Mode == "" {
		log.Printf("[WARN][MATCH] invalid request: %v", err)
		conn.WriteJSON(utils.Response{
			Type:    "find_match_response",
			Success: false,
			Message: invalidRequestMessage,
		})
		return
	}

	username := req.Username
	log.Printf("[INFO][MATCH] matchmaking request: user=%s, mode=%s", username, req.Mode)

	// Check if user already in matchmaking queue
	pendingMu.Lock()
	if pendingPlayers[username] {
		pendingMu.Unlock()
		log.Printf("[WARN][MATCH] user %s already in queue", username)
		conn.WriteJSON(utils.Response{
			Type:    "find_match_response",
			Success: false,
			Message: "Already in queue",
		})
		return
	}
	pendingPlayers[username] = true
	pendingMu.Unlock()

	// Save client connection for future communication
	clientConn := &ClientConnection{Conn: conn}
	clientsMu.Lock()
	clients[username] = clientConn
	clientsMu.Unlock()

	// Create Player instance and register

	user, _ := model.FindUserByUsername(req.Username)

	player := model.NewPlayer(&user, req.Mode)
	model.RegisterConnection(conn, player)

	// Send response to client confirming queue entry
	clientConn.SafeWrite(utils.Response{
		Type:    "find_match_response",
		Success: true,
		Message: "Added to match queue. Waiting for opponent...",
	})

	// Push player to matchmaking queue in a goroutine
	go func() {
		queue, ok := matchQueues[req.Mode]
		if !ok {
			log.Printf("[WARN][MATCH] invalid mode %s for user %s", req.Mode, username)
			clientConn.SafeWrite(utils.Response{
				Type:    "find_match_response",
				Success: false,
				Message: "Invalid game mode",
			})
			return
		}

		select {
		case queue <- player:
		case <-time.After(30 * time.Second):
			log.Printf("[WARN][MATCH] enqueue timeout for user %s", username)
			cleanupUser(username)
			return
		}

		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()

		select {
		case <-player.Matched:
			log.Printf("[INFO][MATCH] user %s matched", username)
		case <-timer.C:
			log.Printf("[WARN][MATCH] matchmaking timeout for user %s", username)
			cleanupUser(username)
			clientConn.SafeWrite(utils.Response{
				Type:    "match_timeout",
				Success: false,
				Message: "Matchmaking timed out. No opponents found.",
			})
		}
	}()

	// Start the matchmaker loop once
	matchmakerOnce.Do(func() {
		startMatchmaker()
		log.Println("[MATCH] matchmaker started")
	})
}

func cleanupUser(username string) {
	pendingMu.Lock()
	delete(pendingPlayers, username)
	pendingMu.Unlock()

	clientsMu.Lock()
	delete(clients, username)
	clientsMu.Unlock()

	log.Printf("[INFO][CLEANUP] removed user %s from queues", username)
}

func startMatchmaker() {
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			for mode, queue := range matchQueues {
				if len(queue) < 2 {
					continue
				}
				player1, player2 := <-queue, <-queue
				log.Printf("[INFO][MATCH] pairing players in mode %s: %s vs %s", mode, player1.User.Username, player2.User.Username)
				if validatePlayers(player1, player2, mode) {
					handleMatch(player1, player2)
				}
			}
		}
	}()
}

func validatePlayers(p1, p2 *model.Player, mode string) bool {
	clientsMu.RLock()
	_, ok1 := clients[p1.User.Username]
	_, ok2 := clients[p2.User.Username]
	clientsMu.RUnlock()

	if !ok1 || !ok2 || p1.User.Username == p2.User.Username {
		if ok1 {
			matchQueues[mode] <- p1
		}
		if ok2 && p1.User.Username != p2.User.Username {
			matchQueues[mode] <- p2
		}
		log.Printf("[WARN][MATCH] validation failed for %s vs %s", p1.User.Username, p2.User.Username)
		return false
	}

	return true
}

func handleMatch(p1, p2 *model.Player) {
	clientsMu.RLock()
	conn1 := clients[p1.User.Username]
	conn2 := clients[p2.User.Username]
	clientsMu.RUnlock()

	// Signal players they've been matched
	p1.Matched <- true
	p2.Matched <- true

	// Create room for matched players
	roomID := utils.GenerateRoomID()
	room := NewRoom(roomID, p1, p2)

	roomsMu.Lock()
	rooms[roomID] = room
	roomsMu.Unlock()

	RegisterRoom(roomID, room)

	log.Printf("[INFO][ROOM] created room %s with players %s and %s", roomID, p1.User.Username, p2.User.Username)

	// Notify both players
	notifyMatchFound(conn1, p2.User.Username, roomID)
	notifyMatchFound(conn2, p1.User.Username, roomID)

	// Remove both players from pending list
	pendingMu.Lock()
	delete(pendingPlayers, p1.User.Username)
	delete(pendingPlayers, p2.User.Username)
	pendingMu.Unlock()
}

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
