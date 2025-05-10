// internal/model/player.go

package model

type Player struct {
	Mana   int               `json:"mana"`
	Towers map[string]*Tower `json:"towers"`
	Troops []*Troop          `json:"troops"`
	Active bool              `json:"active"` 
	User   *User             `json:"user"`  
}

func NewPlayer(user *User, mode string) *Player {
	var troops []*Troop
	if mode != "simple" && mode != "enhanced" {
		return nil
	} else if mode == "simple" {
		troops = getRandomTroops(3)
	} else {
		troops = getRandomTroops(6)
	}

	towers := LoadTower()

	return &Player{
		Mana: 5,
		Towers: map[string]*Tower{
			"king":   towers["King Tower"],
			"guard1": towers["Guard Tower"],
			"guard2": towers["Guard Tower"],
		},
		Troops: troops,
		Active: false,
		User:   user,
	}
}

func (p *Player) ApplyDefenseBoost(percent float64) {
	for _, t := range p.Towers {
		t.IncreaseDefense(percent)
	}
}

func (p *Player) BoostAllTroops() {
	for _, t := range p.Troops {
		t.BoostAttack()
	}
}

func (p *Player) FullyChargeMana() {
	p.Mana = 10
}

func (p *Player) Reset() {
	p.Mana = 5
	p.Towers["king"].Reset()
	p.Towers["guard1"].Reset()
	p.Towers["guard2"].Reset()
	p.Troops = getRandomTroops(3)
}

func (p *Player) DestroyedCount() int {
	count := 0
	for _, t := range p.Towers {
		if t.HP <= 0 {
			count++
		}
	}
	return count
}
