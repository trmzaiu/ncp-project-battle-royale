// internal/model/tower.go

package model

import (
	"encoding/json"
	"os"
)

type Tower struct {
	Type  string  `json:"type"`
	MaxHP int     `json:"max_hp"`
	HP    int     `json:"hp"`
	ATK   int     `json:"atk"`
	DEF   int     `json:"def"`
	CRIT  float64 `json:"crit"`
	EXP   int     `json:"exp"`
}

func LoadTower() map[string]*Tower {
	file, err := os.Open("assets/data/towers.json")
	if err != nil {
		return nil
	}
	defer file.Close()

	var towers []Tower
	if err := json.NewDecoder(file).Decode(&towers); err != nil {
		return nil
	}

	// Convert slice to map for easy access by tower type
	towerMap := make(map[string]*Tower)
	for _, t := range towers {
		towerMap[t.Type] = &t
	}

	return towerMap
}

func (t *Tower) TakeDamage(rawAtk int, attackerLevel int) (actualDamage int, destroyed bool) {
	def := int(float64(t.DEF) * (1 + 0.1*float64(attackerLevel)))
	dmg := rawAtk - def
	if dmg < 0 {
		dmg = 0
	}
	t.HP -= dmg
	if t.HP < 0 {
		t.HP = 0
	}
	return dmg, t.HP == 0
}

func (t *Tower) IncreaseDefense(percent float64) {
	t.DEF = int(float64(t.DEF) * (1 + percent))
}

func (t *Tower) Heal(amount int) {
	t.HP += amount
	if t.HP > t.MaxHP {
		t.HP = t.MaxHP
	}
}

func (t *Tower) Reset() {
	t.HP = 1000
	t.ATK = 300
	t.DEF = 100
	t.CRIT = 0.05
	t.EXP = 100
}
