package game

import (
	"royaka/internal/model"
	"sync"
)

var (
	clients        = make(map[string]*ClientConnection)
	clientsMu      sync.RWMutex

	pendingPlayers = make(map[string]bool)
	pendingMu      sync.RWMutex

	rooms          = make(map[string]*Room)
	roomsMu        sync.RWMutex

	matchQueue        = make(chan *model.Player, 100)
	matchmakerRunning bool

	invalidRequestMessage = "Invalid request"
)
