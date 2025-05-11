// internal/model/tower.go

package model

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"
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

var defaultTowers map[string]*Tower

func init() {
	defaultTowers = LoadTower()
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

func (t *Tower) CounterDamage() int {
	rand.Seed(time.Now().UnixNano())

	baseDamage := t.ATK

	if rand.Float64() < t.CRIT {
		baseDamage = int(float64(baseDamage) * 1.5)
	}

	return baseDamage
}

func (t *Tower) Reset() {
	if def, ok := defaultTowers[t.Type]; ok {
		t.MaxHP = def.MaxHP
		t.HP = def.MaxHP
		t.ATK = def.ATK
		t.DEF = def.DEF
		t.CRIT = def.CRIT
		t.EXP = def.EXP
	} else {
		t.MaxHP = 1000
		t.HP = 1000
		t.ATK = 300
		t.DEF = 100
		t.CRIT = 0.05
		t.EXP = 100
	}
}