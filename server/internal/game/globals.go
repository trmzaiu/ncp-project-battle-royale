package game

import (
	"royaka/internal/model"
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
)
