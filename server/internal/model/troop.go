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
	Name          string  `json:"name"`
	MaxHP         float64 `json:"max_hp"`
	HP            float64 `json:"hp"`
	DMG           float64 `json:"dmg"`
	ATK           float64 `json:"atk"`
	DEF           float64 `json:"def"`
	MANA          int     `json:"mana"`
	CRIT          int     `json:"crit"`
	EXP           int     `json:"exp"`
	Speed         float64 `json:"speed"`
	Range         float64 `json:"range"`
	Type          string  `json:"type"`
	Image         string  `json:"image"`
	Description   string  `json:"description"`
	AOE           bool    `json:"aoe"`
	AttackSpeed   float64 `json:"attack_speed"`
	AggroPriority string  `json:"aggro_priority"`
	Rarity        string  `json:"rarity"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TroopInstance struct {
	ID             string    `json:"id"`
	Template       *Troop    `json:"template"`
	TypeEntity     string    `json:"type_entity"`
	Owner          string    `json:"owner"`
	Position       Position  `json:"position"`
	TargetID       string    `json:"target_id"`
	TargetType     string    `json:"target_type"`
	IsDead         bool      `json:"is_dead"`
	LastAttackTime time.Time `json:"last_attack"`
}

func (t *TroopInstance) GetID() string {
	return t.ID
}

func (t *TroopInstance) GetOwner() string {
	return t.Owner
}

func (t *TroopInstance) GetType() string {
	return t.TypeEntity
}

func (t *TroopInstance) GetPosition() Position {
	return t.Position
}

func (t *TroopInstance) IsAlive() bool {
	return !t.IsDead
}

func LoadTroop() ([]Troop, error) {
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
	templates, err := LoadTroop()
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

func createTroopInstances(templates []*Troop, owner string) []*TroopInstance {
	instances := make([]*TroopInstance, 0, len(templates))
	for _, troop := range templates {
		instance := &TroopInstance{
			ID:             uuid.New().String(),
			Template:       troop,
			TypeEntity:     "troop",
			Owner:          owner,
			LastAttackTime: time.Now(),
		}
		instances = append(instances, instance)
	}
	return instances
}

func shuffleTroops(troops []*Troop) []*Troop {
	shuffled := make([]*Troop, len(troops))
	copy(shuffled, troops)
	for i := len(shuffled) - 1; i > 0; i-- {
		j, err := cryptoRandInt(int64(i + 1))
		if err != nil {
			continue
		}
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

func (t *Troop) CalculateDamage(level int) (float64, bool) {
	baseAtk := t.ATK * (1 + 0.1*float64(level))

	// Use crypto/rand for crit calculation
	critRoll, err := cryptoRandInt(100)
	if err != nil {
		return baseAtk, false
	}
	isCrit := critRoll < int64(t.CRIT)

	if isCrit {
		baseAtk *= 1.5
	}

	return baseAtk, isCrit
}

func (t *Troop) CalculateHeal(level int) (float64, bool) {
	baseHp := t.ATK * (1 + 0.1*float64(level))

	// Use crypto/rand for crit calculation
	critRoll, err := cryptoRandInt(100)
	if err != nil {
		return baseHp, false
	}
	isCrit := critRoll < int64(t.CRIT)

	if isCrit {
		baseHp *= 1.5
	}

	return baseHp, isCrit
}

func (t *Troop) BoostAttack() {
	t.ATK = float64(t.ATK) * 1.5
}

func (t *Troop) FortifyHP(amount float64) {
	t.HP += amount
	if t.HP > t.MaxHP {
		t.HP = t.MaxHP
	}
}
