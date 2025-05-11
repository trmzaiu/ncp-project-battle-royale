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
	rooms   = make(map[string]*model.Room)
	roomsMu sync.RWMutex

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
		User *model.User `json:"user"`
		Mode string      `json:"mode"`
	}

	if err := json.Unmarshal(data, &req); err != nil || req.User == nil || req.Mode == "" {
		log.Printf("Failed to parse start_game request: %v", err)
		conn.WriteJSON(utils.Response{Type: "start_game_response", Success: false, Message: "Invalid request"})
		return
	}

	// Check if player is already in the queue
	pendingPlayersMu.Lock()
	if pendingPlayers[req.User.Username] {
		pendingPlayersMu.Unlock()
		conn.WriteJSON(utils.Response{Type: "start_game_response", Success: false, Message: "Already in queue"})
		return
	}

	// Mark this player as pending
	pendingPlayers[req.User.Username] = true
	pendingPlayersMu.Unlock()

	// Create or update client connection with mutex protection
	clientConn := &ClientConnection{Conn: conn, Mu: sync.Mutex{}}

	clientsMu.Lock()
	clients[req.User.Username] = clientConn
	clientsMu.Unlock()

	player := model.NewPlayer(req.User, req.Mode)

	// Send response to client
	clientConn.SafeWrite(utils.Response{
		Type:    "start_game_response",
		Success: true,
		Message: "Added to match queue. Waiting for opponent...",
	})

	// Add player to queue in a separate goroutine
	go func() {
		log.Printf("Player %s added to match queue", req.User.Username)
		matchQueue <- player

		if !matchmakerRunning {
			log.Println("Starting matchmaker...")
			startMatchmaker()
		}
	}()
}

func HandleGetGameInfo(conn *websocket.Conn, data json.RawMessage) {
	var req struct {
		RoomID   string `json:"room_id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(data, &req); err != nil || req.RoomID == "" || req.Username == "" {
		conn.WriteJSON(utils.Response{Type: "game_info_response", Success: false, Message: "Invalid request"})
		return
	}

	roomsMu.RLock()
	room, exists := rooms[req.RoomID]
	roomsMu.RUnlock()

	if !exists {
		conn.WriteJSON(utils.Response{Type: "game_info_response", Success: false, Message: "Room not found"})
		return
	}

	var currentUser, opponent *model.Player

	if room.Player1.User.Username == req.Username {
		currentUser = room.Player1
		opponent = room.Player2
	} else if room.Player2.User.Username == req.Username {
		currentUser = room.Player2
		opponent = room.Player1
	}

	if currentUser == nil || opponent == nil {
		conn.WriteJSON(utils.Response{
			Type:    "game_info_response",
			Success: false,
			Message: "Could not identify player in this room",
		})
		return
	}

	conn.WriteJSON(utils.Response{
		Type:    "game_info_response",
		Success: true,
		Message: "Game info loaded",
		Data: map[string]interface{}{
			"user":     currentUser,
			"opponent": opponent,
		},
	})
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
			conn1, ok1 := clients[player1.User.Username]
			conn2, ok2 := clients[player2.User.Username]
			clientsMu.Unlock()

			if !ok1 || !ok2 || player1.User.Username == player2.User.Username {
				// Put valid players back in queue and retry
				if ok1 {
					matchQueue <- player1
				}
				if ok2 && player1.User.Username != player2.User.Username {
					matchQueue <- player2
				}
				continue
			}

			// Start match between the two players
			startMatch(player1, player2, conn1, conn2)

			// Remove players from pending list
			pendingPlayersMu.Lock()
			delete(pendingPlayers, player1.User.Username)
			delete(pendingPlayers, player2.User.Username)
			pendingPlayersMu.Unlock()
		}
	}()
}

func startMatch(p1, p2 *model.Player, conn1, conn2 *ClientConnection) {
	log.Printf("Match started between %s and %s", p1.User.Username, p2.User.Username)

	roomID := generateRoomID()
	room := model.NewRoom(roomID, p1, p2)

	roomsMu.Lock()
	rooms[roomID] = room
	roomsMu.Unlock()

	log.Printf("Created room with ID: %s for players %s and %s", roomID, p1.User.Username, p2.User.Username)

	matchInfo1 := map[string]interface{}{
		"room_id":  roomID,
		"opponent": p2.User.Username,
	}

	matchInfo2 := map[string]interface{}{
		"room_id":  roomID,
		"opponent": p1.User.Username,
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

func startGame(room *model.Room, conn1, conn2 *ClientConnection) {
	var req struct {
		Mode string `json:"mode"`
	}

	var isEnhanced bool

	if req.Mode == "simple" {
		isEnhanced = false
	} else if req.Mode == "enhanced" {
		isEnhanced = true
	} else {
		conn1.SafeWrite(utils.Response{
			Type:    "error",
			Success: false,
			Message: "Invalid game mode",
		})
		conn2.SafeWrite(utils.Response{
			Type:    "error",
			Success: false,
			Message: "Invalid game mode",
		})
		return
	}

	game := NewGame(room.Player1, room.Player2, isEnhanced)

	conn1.SafeWrite(utils.Response{
		Type:    "game_started",
		Success: true,
		Message: "Game has started!",
		Data: map[string]interface{}{
			"player1": room.Player1.User.Username,
			"player2": room.Player2.User.Username,
		},
	})

	conn2.SafeWrite(utils.Response{
		Type:    "game_started",
		Success: true,
		Message: "Game has started!",
		Data: map[string]interface{}{
			"player1": room.Player1.User.Username,
			"player2": room.Player2.User.Username,
		},
	})

	go handleGameTurns(game, conn1, conn2)
}

func handleGameTurns(game *Game, conn1, conn2 *ClientConnection) {
	for {
		currentPlayer := game.CurrentPlayer()
		if currentPlayer == nil {
			log.Println("Enhanced mode or invalid current player.")
			break
		}

		var connCurrent, connOpponent *ClientConnection
		if currentPlayer.User.Username == game.Player1.User.Username {
			connCurrent = conn1
			connOpponent = conn2
		} else {
			connCurrent = conn2
			connOpponent = conn1
		}

		connCurrent.SafeWrite(utils.Response{
			Type:    "player_turn",
			Success: true,
			Message: fmt.Sprintf("It's your turn, %s!", currentPlayer.User.Username),
		})

		var action struct {
			TroopName string `json:"troop"`
			TowerType string `json:"target"`
		}
		if err := connCurrent.Conn.ReadJSON(&action); err != nil {
			log.Println("Error reading action:", err)
			break
		}

		var troop *model.Troop
		for i := range currentPlayer.Troops {
			if currentPlayer.Troops[i].Name == action.TroopName && currentPlayer.Troops[i].HP > 0 {
				troop = currentPlayer.Troops[i]
				break
			}
		}
		if troop == nil {
			connCurrent.SafeWrite(utils.Response{
				Type:    "turn_result",
				Success: false,
				Message: "Invalid or dead troop selected.",
			})
			continue
		}

		result := game.PlayTurn(currentPlayer, troop, action.TowerType)

		connCurrent.SafeWrite(utils.Response{
			Type:    "turn_result",
			Success: true,
			Message: result,
		})
		connOpponent.SafeWrite(utils.Response{
			Type:    "opponent_turn",
			Success: true,
			Message: fmt.Sprintf("%s's move: %s", currentPlayer.User.Username, result),
		})

		winner := game.CheckWinner()
		if winner != "" {
			conn1.SafeWrite(utils.Response{Type: "game_over", Success: true, Message: winner})
			conn2.SafeWrite(utils.Response{Type: "game_over", Success: true, Message: winner})
			break
		}
	}
}

func generateRoomID() string {
	return fmt.Sprintf("room-%d", time.Now().UnixNano())
}
