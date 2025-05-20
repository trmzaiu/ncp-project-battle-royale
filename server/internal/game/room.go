// internal/model/room.go

package game

import (
	"royaka/internal/model"
	"sync"
)

type Room struct {
	ID      string
	Player1 *model.Player
	Player2 *model.Player
	Game    *Game
	mu      sync.Mutex
}

var (
	roomRegistry = make(map[string]*Room)
	roomLock     sync.RWMutex
)

func NewRoom(id string, p1, p2 *model.Player, mode string) *Room {
	return &Room{
		ID:      id,
		Player1: p1,
		Player2: p2,
		Game:    NewGame(p1, p2, mode),
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

func GetRoomIDByUsername(username string) string {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	for id, room := range rooms {
		if (room.Player1 != nil && room.Player1.User.Username == username) ||
			(room.Player2 != nil && room.Player2.User.Username == username) {
			return id
		}
	}
	return ""
}
