// internal/model/room.go

package model

type Room struct {
	ID       string  `json:"id"`
	Player1  *Player `json:"player1"`
	Player2  *Player `json:"player2"`
	IsActive bool    `json:"is_active"`
}

func NewRoom(id string, p1, p2 *Player) *Room {
	return &Room{
		ID:       id,
		Player1:  p1,
		Player2:  p2,
		IsActive: true,
	}
}

