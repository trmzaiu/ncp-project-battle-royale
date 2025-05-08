// internal/player/tower.go

package player

type Tower struct {
	Type  string  `json:"type"`
	HP    int     `json:"hp"`
	ATK   int     `json:"atk"`
	DEF   int     `json:"def"`
	CRIT  float64 `json:"crit"` // percentage 0.05 = 5%
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
