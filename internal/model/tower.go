// internal/model/tower.go

package model

type Tower struct {
	Type  string  `json:"type"`
	HP    int     `json:"hp"`
	ATK   int     `json:"atk"`
	DEF   int     `json:"def"`
	CRIT  float64 `json:"crit"` 
	EXP   int     `json:"exp"`
}

func NewTower(towerType string) *Tower {
	switch towerType {
	case "King Tower":
		return &Tower{Type: towerType, HP: 2000, ATK: 500, DEF: 300, CRIT: 0.1, EXP: 200}
	default:
		return &Tower{Type: towerType, HP: 1000, ATK: 300, DEF: 100, CRIT: 0.05, EXP: 100}
	}
}

func (t *Tower) TakeDamage(amount int) {
	t.HP -= amount
	if t.HP < 0 {
		t.HP = 0
	}
}

func AttackTower(troop *Troop, target *Tower) int {
	dmg := troop.DamageTo(target)
	target.TakeDamage(dmg)
	return dmg
}

func HealLowestTower(towers map[string]*Tower) {
	var lowest *Tower
	for _, t := range towers {
		if t.HP > 0 && (lowest == nil || t.HP < lowest.HP) {
			lowest = t
		}
	}
	if lowest != nil {
		lowest.HP += 300
	}
}