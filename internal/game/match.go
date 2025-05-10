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
	// Connection management
	clients   = make(map[string]*ClientConnection)
	clientsMu sync.Mutex

	// Game matching
	matchQueue       = make(chan *model.Player, 100)
	pendingPlayers   = make(map[string]bool)
	pendingPlayersMu sync.Mutex
	
	// Room management
	rooms      = make(map[string]*model.Room)
	roomsMu    sync.RWMutex
	
	matchmakerRunning bool
)

// ClientConnection wraps a websocket connection with a mutex for thread safety
type ClientConnection struct {
	Conn *websocket.Conn
	Mu   sync.Mutex // Mutex to protect writes to the connection
}

// SafeWrite sends a JSON message to the client with mutex protection
func (c *ClientConnection) SafeWrite(data interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Conn.WriteJSON(data)
}

// HandleStartGame handles player joining the match queue
func HandleStartGame(conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		Username string `json:"username"`
	}

	if err := json.Unmarshal(data, &req); err != nil || req.Username == "" {
		conn.WriteJSON(utils.Response{Type: "start_game_response", Success: false, Message: "Invalid username"})
		return
	}

	// Check if player is already in the queue
	pendingPlayersMu.Lock()
	if pendingPlayers[req.Username] {
		pendingPlayersMu.Unlock()
		conn.WriteJSON(utils.Response{Type: "start_game_response", Success: false, Message: "Already in queue"})
		return
	}
	
	// Mark this player as pending
	pendingPlayers[req.Username] = true
	pendingPlayersMu.Unlock()

	// Create or update client connection with mutex protection
	clientConn := &ClientConnection{Conn: conn, Mu: sync.Mutex{}}
	
	clientsMu.Lock()
	clients[req.Username] = clientConn
	clientsMu.Unlock()

	player := model.NewPlayer(req.Username)

	// Send response to client
	clientConn.SafeWrite(utils.Response{
		Type:    "start_game_response",
		Success: true,
		Message: "Added to match queue. Waiting for opponent...",
	})

	// Add player to queue in a separate goroutine
	go func() {
		log.Printf("Player %s added to match queue", req.Username)
		matchQueue <- player

		if !matchmakerRunning {
			log.Println("Starting matchmaker...")
			startMatchmaker()
		}
	}()
}

func startMatchmaker() {
	matchmakerRunning = true
	go func() {
		for {
			// Wait for at least 2 players in queue
			if len(matchQueue) < 2 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			
			// Get two players
			player1 := <-matchQueue
			player2 := <-matchQueue
			
			// Make sure they're not the same player or we still have their connections
			clientsMu.Lock()
			conn1, ok1 := clients[player1.Username]
			conn2, ok2 := clients[player2.Username]
			clientsMu.Unlock()
			
			if !ok1 || !ok2 || player1.Username == player2.Username {
				// Put valid players back in queue and retry
				if ok1 {
					matchQueue <- player1
				}
				if ok2 && player1.Username != player2.Username {
					matchQueue <- player2
				}
				continue
			}
			
			// Start match between the two players
			startMatch(player1, player2, conn1, conn2)
			
			// Remove players from pending list
			pendingPlayersMu.Lock()
			delete(pendingPlayers, player1.Username)
			delete(pendingPlayers, player2.Username)
			pendingPlayersMu.Unlock()
		}
	}()
}

func startMatch(p1, p2 *model.Player, conn1, conn2 *ClientConnection) {
	log.Printf("Match started between %s and %s", p1.Username, p2.Username)

	roomID := generateRoomID()
	room := model.NewRoom(roomID, p1, p2)
	
	roomsMu.Lock()
	rooms[roomID] = room
	roomsMu.Unlock()

	matchInfo1 := map[string]interface{}{
		"room_id":  roomID,
		"opponent": p2.Username,
	}
	
	matchInfo2 := map[string]interface{}{
		"room_id":  roomID,
		"opponent": p1.Username,
	}
	
	// Send match found notifications safely
	err1 := conn1.SafeWrite(utils.Response{
		Type:    "match_found",
		Success: true,
		Message: "Match found!",
		Data:    matchInfo1,
	})
	
	err2 := conn2.SafeWrite(utils.Response{
		Type:    "match_found",
		Success: true,
		Message: "Match found!",
		Data:    matchInfo2,
	})
	
	if err1 != nil || err2 != nil {
		log.Printf("Error sending match notifications: %v, %v", err1, err2)
	}
}

func generateRoomID() string {
	return fmt.Sprintf("room-%d", time.Now().UnixNano())
}