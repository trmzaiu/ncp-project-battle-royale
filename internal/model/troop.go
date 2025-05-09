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
	HP      int     `json:"hp"`
	ATK     int     `json:"atk"`
	DEF     int     `json:"def"`
	MANA    int     `json:"mana"`
	EXP     int     `json:"exp"`
	CRIT    float64 `json:"crit"`
	Special string  `json:"special"`
}

func NewTroop(name string) *Troop {
	switch name {
	case "Pawn":
		return &Troop{name, 50, 150, 100, 3, 5, 0, ""}
	case "Bishop":
		return &Troop{name, 100, 200, 150, 4, 10, 0, ""}
	case "Rook":
		return &Troop{name, 250, 200, 200, 5, 25, 0, ""}
	case "Knight":
		return &Troop{name, 200, 300, 150, 5, 25, 0, ""}
	case "Prince":
		return &Troop{name, 500, 400, 300, 6, 50, 0, ""}
	case "Queen":
		return &Troop{name, 0, 0, 0, 5, 30, 0, "Heal"}
	default:
		return &Troop{name, 100, 100, 100, 3, 5, 0, ""}
	}
}

func (t *Troop) IsCrit() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64() < 0.1 // 10% base CRIT chance
}

func (t *Troop) DamageTo(target *Tower) int {
	crit := t.IsCrit()
	atk := t.ATK
	if crit {
		atk = int(float64(atk) * 1.2)
	}
	dmg := atk - target.DEF
	if dmg < 0 {
		return 0
	}
	return dmg
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

func getRandomTroops(n int) ([]*Troop, error) {
	templates, err := loadTroop()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(templates), func(i, j int) {
		templates[i], templates[j] = templates[j], templates[i]
	})

	selected := make([]*Troop, 0, n)
	for i := 0; i < n && i < len(templates); i++ {
		t := templates[i]
		selected = append(selected, &Troop{
			Name: t.Name,
			ATK:  t.ATK,
			DEF:  t.DEF,
			CRIT: t.CRIT,
		})
	}
	return selected, nil
}
