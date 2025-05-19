package game

import (
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"
	"sync"
)

var (
	clients   = make(map[string]*ClientConnection)
	clientsMu sync.RWMutex

	pendingPlayers = make(map[string]bool)
	pendingMu      sync.RWMutex

	rooms   = make(map[string]*Room)
	roomsMu sync.RWMutex

	matchQueues = map[string]chan *model.Player{
		"simple":   make(chan *model.Player, 100),
		"enhanced": make(chan *model.Player, 100),
	}
	matchmakerOnce    sync.Once

	invalidRequestMessage = "Invalid request"
	roomRequestMessage    = "Room not found"
	manaRequestMessage = "Not enough mana!"
)

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
