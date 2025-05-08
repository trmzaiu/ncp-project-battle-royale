// internal/player/player.go

package player

import (

)

type Player struct {
	Username string                 `json:"username"`
	Password string                 `json:"password"`
	EXP      int                    `json:"exp"`
	Level    int                    `json:"level"`
	Mana     int                    `json:"mana"`
	Towers   map[string]*Tower `json:"towers"`
	Troops   []*Troop          `json:"troops"`
	Active   bool                   `json:"active"`
}

func NewPlayer(username, password string) *Player {
	return &Player{
		Username: username,
		Password: password,
		EXP:      0,
		Level:    1,
		Mana:     5,
		Towers: map[string]*Tower{
			"king":   NewTower("King Tower"),
			"guard1": NewTower("Guard Tower"),
			"guard2": NewTower("Guard Tower"),
		},
		Troops: make([]*Troop, 0),
		Active: false,
	}
}

func (p *Player) GainEXP(amount int) {
	p.EXP += amount
	for p.EXP >= p.requiredEXP() {
		p.Level++
		p.EXP -= p.requiredEXP()
	}
}

func (p *Player) requiredEXP() int {
	return 100 + (p.Level-1)*10
}

func (p *Player) ManaRegen() {
	if p.Mana < 10 {
		p.Mana++
	}
}
