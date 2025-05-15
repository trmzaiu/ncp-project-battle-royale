package model

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"os"
)

type Troop struct {
	Name        string  `json:"name"`
	MaxHP       int     `json:"max_hp"`
	HP          int     `json:"hp"`
	ATK         int     `json:"atk"`
	DEF         int     `json:"def"`
	MANA        int     `json:"mana"`
	EXP         int     `json:"exp"`
	CRIT        float64 `json:"crit"`
	Special     string  `json:"special"`
	Icon        string  `json:"icon"`
	Description string  `json:"description"`
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

func cryptoRandInt(max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

func getRandomTroops(n int) []*Troop {
	templates, err := loadTroop()
	if err != nil {
		return nil
	}

	// Fisher–Yates Shuffle with crypto/rand
	for i := len(templates) - 1; i > 0; i-- {
		j64, err := cryptoRandInt(int64(i + 1))
		if err != nil {
			continue
		}
		j := int(j64)
		templates[i], templates[j] = templates[j], templates[i]
	}

	selected := make([]*Troop, 0, n)
	for i := 0; i < n && i < len(templates); i++ {
		t := templates[i]
		selected = append(selected, &Troop{
			Name:        t.Name,
			ATK:         t.ATK,
			DEF:         t.DEF,
			CRIT:        t.CRIT,
			MaxHP:       t.MaxHP,
			HP:          t.MaxHP,
			MANA:        t.MANA,
			EXP:         t.EXP,
			Special:     t.Special,
			Icon:        t.Icon,
			Description: t.Description,
		})
	}
	return selected
}

func (t *Troop) CalculateDamage(level int) (int, bool) {
	baseAtk := float64(t.ATK) * (1 + 0.1*float64(level))

	// Use crypto/rand for crit calculation
	critRoll, err := cryptoRandInt(100)
	if err != nil {
		return int(baseAtk), false
	}
	isCrit := critRoll < int64(t.CRIT)

	if isCrit {
		baseAtk *= 1.5
	}

	return int(baseAtk), isCrit
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
			"icon":  t.Icon,
			"desc":  t.Description,
		})
	}
	return troops
}
