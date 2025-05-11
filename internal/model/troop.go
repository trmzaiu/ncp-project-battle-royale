// internal/model/troop.go

package model

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
)

type Troop struct {
	Name    string  `json:"name"`
	MaxHP   int     `json:"max_hp"`
	HP      int     `json:"hp"`
	ATK     int     `json:"atk"`
	DEF     int     `json:"def"`
	MANA    int     `json:"mana"`
	EXP     int     `json:"exp"`
	CRIT    float64 `json:"crit"`
	Special string  `json:"special"`
}

func loadTroop() ([]Troop, error) {
	file, err := os.Open("assets/data/troops.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var templates []Troop
	if err := json.NewDecoder(file).Decode(&templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func getRandomTroops(n int) []*Troop {
	templates, err := loadTroop()
	if err != nil {
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(templates), func(i, j int) {
		templates[i], templates[j] = templates[j], templates[i]
	})

	selected := make([]*Troop, 0, n)
	for i := 0; i < n && i < len(templates); i++ {
		t := templates[i]
		selected = append(selected, &Troop{
			Name:  t.Name,
			ATK:   t.ATK,
			DEF:   t.DEF,
			CRIT:  t.CRIT,
			MaxHP: t.MaxHP,
			HP:    t.MaxHP,
		})
	}
	return selected
}

func (t *Troop) CalculateDamage(level int, critEnabled bool) int {
	atk := float64(t.ATK) * (1 + 0.1*float64(level))
	if critEnabled && IsCriticalHit(int(t.CRIT)) {
		atk *= 1.2
	}
	return int(atk)
}

func IsCriticalHit(critChance int) bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(100) < critChance
}

func (t *Troop) BoostAttack() {
	t.ATK = int(float64(t.ATK) * 1.5)
}

func (t *Troop) FortifyHP(amount int) {
	t.HP += amount
	if t.HP > t.MaxHP {
		t.HP = t.MaxHP
	}
}

func (p *Player) TowerStatus() map[string]int {
	status := make(map[string]int)
	for k, v := range p.Towers {
		status[k] = v.HP
	}
	return status
}

func (p *Player) TroopStatus() []map[string]interface{} {
	var troops []map[string]interface{}
	for _, t := range p.Troops {
		troops = append(troops, map[string]interface{}{
			"name":  t.Name,
			"hp":    t.HP,
			"mana":  t.MANA,
			"skill": t.Special,
		})
	}
	return troops
}

