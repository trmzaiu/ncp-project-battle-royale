// internal/model/room.go

package game

import (
	"royaka/internal/model"
	"sync"
)

type Room struct {
	ID      string        `json:"id"`
	Player1 *model.Player `json:"player1"`
	Player2 *model.Player `json:"player2"`
	Game    *Game         `json:"game"`
}

var (
	roomRegistry = make(map[string]*Room)
	roomLock     sync.RWMutex
)

func NewRoom(id string, p1, p2 *model.Player) *Room {
	return &Room{
		ID:      id,
		Player1: p1,
		Player2: p2,
		Game:    NewGame(p1, p2, false),
	}
}

func RegisterRoom(roomID string, room *Room) {
	roomLock.Lock()
	defer roomLock.Unlock()
	roomRegistry[roomID] = room
}

func GetRoom(roomID string) *Room {
	roomLock.RLock()
	defer roomLock.RUnlock()
	return roomRegistry[roomID]
}

func RemoveRoom(roomID string) {
	roomLock.Lock()
	defer roomLock.Unlock()
	delete(roomRegistry, roomID)
}
