package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
	},
}

// Handle incoming WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	player := addPlayer(id, conn)
	log.Printf("✅ Player %s connected", id)
	conn.WriteMessage(websocket.TextMessage, []byte("Welcome "+id+"!"))

	// Handle incoming messages from the player
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Player %s disconnected: %v", id, err)
			removePlayer(id)
			return
		}

		handleCommand(player, string(msg)) // Process the command
	}
}

var playersLock sync.Mutex

// Remove player from the game
func removePlayer(id string) {
	playersLock.Lock()
	delete(players, id)
	playersLock.Unlock()
}

// handleCommand processes the message sent by the player
func handleCommand(player *Player, cmd string) {
	switch cmd {
	case "move up":
		movePlayer(player, "up")
	case "move down":
		movePlayer(player, "down")
	case "move left":
		movePlayer(player, "left")
	case "move right":
		movePlayer(player, "right")
	default:
		log.Printf("Unknown command from player %s: %s", player.ID, cmd)
	}
}
