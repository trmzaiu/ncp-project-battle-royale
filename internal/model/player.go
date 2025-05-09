// internal/model/player.go

package model

type Player struct {
	Username string            `json:"username"`
	Mana     int               `json:"mana"`
	Towers   map[string]*Tower `json:"towers"`
	Troops   []*Troop          `json:"troops"`
	Active   bool              `json:"active"` // Is currently in a game
}

func NewPlayer(username string) *Player {
	return &Player{
		Username: username,
		Mana: 5,
		Towers: map[string]*Tower{
			"king":   NewTower("King Tower"),
			"guard1": NewTower("Guard Tower"),
			"guard2": NewTower("Guard Tower"),
		},
		Troops: make([]*Troop, 0),
		Active: false,
	}
}

func (p *Player) ManaRegen() {
	if p.Mana < 10 {
		p.Mana++
	}
}

func (p *Player) Reset() {
	// Reset tower HP to their default values
	p.Towers["king"].HP = 2000   // King Tower has 2000 HP
	p.Towers["guard1"].HP = 1000 // Guard Towers have 1000 HP
	p.Towers["guard2"].HP = 1000
	p.Mana = 5
}
