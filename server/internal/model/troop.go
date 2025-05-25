package model

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"royaka/internal/utils"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ==== STRUCTS ====

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
	AttackSpeed   float64 `json:"attack_speed"`
	AggroPriority string  `json:"aggro_priority"`
	Rarity        string  `json:"rarity"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TroopInstance struct {
	ID             string       `json:"id"`
	Template       *Troop       `json:"template"`
	TypeEntity     string       `json:"type_entity"`
	Owner          string       `json:"owner"`
	Position       Position     `json:"position"`
	IsDead         bool         `json:"is_dead"`
	LastAttackTime time.Time    `json:"last_attack"`
	Mutex          sync.RWMutex `json:"-"`
}

// -------- Getters --------

func (t *TroopInstance) GetID() string    { return t.ID }
func (t *TroopInstance) GetOwner() string { return t.Owner }
func (t *TroopInstance) GetType() string  { return t.TypeEntity }
func (t *TroopInstance) GetPosition() Position { return t.Position }

func (t *TroopInstance) IsAlive() bool {
	return !t.IsDead && t.Template.HP > 0
}

func (p Position) String() string {
	x := int(math.Floor(p.X))
	y := int(math.Floor(p.Y))
	return fmt.Sprintf("%d_%d", x, y)
}

// -------- Loading & Random Utils --------

// Load troop templates from JSON file
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

// Shuffle troop slice securely
func shuffleTroops(troops []*Troop) []*Troop {
	shuffled := make([]*Troop, len(troops))
	copy(shuffled, troops)

	for i := len(shuffled) - 1; i > 0; i-- {
		j, err := utils.CryptoRandInt(int64(i + 1))
		if err != nil {
			continue
		}
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}

// Get random n troops, HP reset to MaxHP
func getRandomTroops(n int) []*Troop {
	templates, err := LoadTroop()
	if err != nil {
		return nil
	}

	shuffled := shuffleTroops(pointerizeTroops(templates))
	if n > len(shuffled) {
		n = len(shuffled)
	}

	selected := make([]*Troop, n)
	for i := 0; i < n; i++ {
		t := *shuffled[i] // copy struct
		t.HP = t.MaxHP    // reset HP full
		selected[i] = &t
	}
	return selected
}

// helper to convert slice of Troop structs to slice of pointers
func pointerizeTroops(ts []Troop) []*Troop {
	result := make([]*Troop, len(ts))
	for i := range ts {
		result[i] = &ts[i]
	}
	return result
}

// Create TroopInstance slice from troop templates
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

func (troop *TroopInstance) InAttackRange(targetPos Position) bool {
    dx := troop.Position.X - targetPos.X
    dy := troop.Position.Y - targetPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= troop.Template.Range
}

// -------- Combat Calculations --------

// Calculate damage with crit chance (level used for scaling)
func (t *Troop) CalculateDamage(level int) (float64, bool) {
	baseAtk := t.ATK * (1 + 0.1*float64(level))

	critRoll, err := utils.CryptoRandInt(100)
	if err != nil {
		return baseAtk, false
	}
	isCrit := critRoll < int64(t.CRIT)

	if isCrit {
		baseAtk *= 1.5
	}

	return baseAtk, isCrit
}

// Calculate heal with crit chance (level used for scaling)
func (t *Troop) CalculateHeal(level int) (float64, bool) {
	baseHeal := t.MaxHP / 3 * (1 + 0.1*float64(level))

	critRoll, err := utils.CryptoRandInt(100)
	if err != nil {
		return baseHeal, false
	}
	isCrit := critRoll < int64(t.CRIT)

	if isCrit {
		baseHeal *= 1.5
	}

	return baseHeal, isCrit
}

// Boost attack by 50%
func (t *Troop) BoostAttack() {
	t.ATK *= 1.5
}

// Heal (Fortify HP) with cap at MaxHP
func (t *Troop) FortifyHP(amount float64) {
	t.HP += amount
	if t.HP > t.MaxHP {
		t.HP = t.MaxHP
	}
}
