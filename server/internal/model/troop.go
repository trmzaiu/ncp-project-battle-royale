package model

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
)

type Troop struct {
	Name            string  `json:"name"`
	MaxHP           int     `json:"max_hp"`
	HP              int     `json:"hp"`
	ATK             int     `json:"atk"`
	DEF             int     `json:"def"`
	MANA            int     `json:"mana"`
	CRIT            int     `json:"crit"`
	Speed           float64 `json:"speed"`
	Range           float64 `json:"range"`
	Type            string  `json:"type"`
	Card            string  `json:"card"`
	Image           string  `json:"image"`
	Description     string  `json:"description"`
	AOE             bool    `json:"aoe"`
	AttackSpeed     float64 `json:"attack_speed"`
	AggroPriority   int     `json:"aggro_priority"`
	ProjectileSpeed float64 `json:"projectile_speed"`
}

type TroopInstance struct {
	ID             string    `json:"id"`
	Template       *Troop    `json:"template"`
	OwnerID        string    `json:"owner_id"`
	X, Y           float64   `json:"x", "y"`
	TargetID       string    `json:"target_id"`
	TargetType     string    `json:"target_type"`
	IsDead         bool      `json:"is_dead"`
	LastAttackTime time.Time `json:"last_attack"`
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

	// Shuffle
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
		copy := t
		copy.HP = t.MaxHP
		selected = append(selected, &copy)
	}
	return selected
}

func createTroopInstances(templates []*Troop, ownerID string) []*TroopInstance {
	instances := make([]*TroopInstance, 0, len(templates))
	for _, t := range templates {
		instance := &TroopInstance{
			ID:             uuid.New().String(),
			Template:       t,
			OwnerID:        ownerID,
			LastAttackTime: time.Now(),
		}
		instances = append(instances, instance)
	}
	return instances
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

func (t *Troop) CalculateHeal(level int) (int, bool) {
	baseHp := float64(t.ATK) * (1 + 0.1*float64(level))

	// Use crypto/rand for crit calculation
	critRoll, err := cryptoRandInt(100)
	if err != nil {
		return int(baseHp), false
	}
	isCrit := critRoll < int64(t.CRIT)

	if isCrit {
		baseHp *= 1.5
	}

	return int(baseHp), isCrit
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
