package model

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ==== STRUCTS ====

type Tower struct {
	Type        string  `json:"type"`
	MaxHP       float64 `json:"max_hp"`
	HP          float64 `json:"hp"`
	ATK         float64 `json:"atk"`
	DEF         float64 `json:"def"`
	CRIT        float64 `json:"crit"`
	EXP         int     `json:"exp"`
	Range       float64 `json:"range"`
	AttackSpeed float64 `json:"attack_speed"`
}

type Area struct {
	TopLeft     Position `json:"top_left"`
	BottomRight Position `json:"bottom_right"`
}

type TowerInstance struct {
	ID             string       `json:"id"`
	Template       *Tower       `json:"template"`
	TypeEntity     string       `json:"type_entity"`
	Owner          string       `json:"owner"`
	Area           Area         `json:"area"`
	IsDestroyed    bool         `json:"is_destroyed"`
	LastAttackTime time.Time    `json:"last_attack"`
	Mutex          sync.RWMutex `json:"-"`
}

// -------- Getters --------

func (t *TowerInstance) GetID() string    { return t.ID }
func (t *TowerInstance) GetOwner() string { return t.Owner }
func (t *TowerInstance) GetType() string  { return t.TypeEntity }
func (t *TowerInstance) GetPosition() Position {
	// Return center of tower area
	return Position{
		X: (t.Area.TopLeft.X + t.Area.BottomRight.X) / 2,
		Y: (t.Area.TopLeft.Y + t.Area.BottomRight.Y) / 2,
	}
}

func (t *TowerInstance) IsAlive() bool {
	return !t.IsDestroyed && t.Template.HP > 0
}

// ---------- Initialization ----------

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

	towerMap := make(map[string]*Tower)
	for _, t := range towers {
		towerMap[t.Type] = &t
	}
	return towerMap
}

// ---------- Tower Creation ----------

func CreateTowerInstances(towers map[string]*Tower, owner string, isPlayer1 bool) []*TowerInstance {
	instances := []*TowerInstance{}
	for _, tower := range towers {
		instance := &TowerInstance{
			ID:             uuid.New().String(),
			Template:       tower,
			TypeEntity:     "tower",
			Owner:          owner,
			IsDestroyed:    false,
			Area:           GetTowerArea(tower.Type, isPlayer1),
			LastAttackTime: time.Now(),
		}
		instances = append(instances, instance)
	}
	return instances
}

func GetTowerArea(towerType string, isPlayer1 bool) Area {
	switch towerType {
	case "king":
		if isPlayer1 {
			return Area{
				TopLeft:     Position{9, 0},
				BottomRight: Position{12, 3},
			}
		}
		return Area{
			TopLeft:     Position{9, 18},
			BottomRight: Position{12, 21},
		}
	case "guard1":
		if isPlayer1 {
			return Area{
				TopLeft:     Position{3, 2},
				BottomRight: Position{5, 4},
			}
		}
		return Area{
			TopLeft:     Position{3, 17},
			BottomRight: Position{5, 19},
		}

	case "guard2":
		if isPlayer1 {
			return Area{
				TopLeft:     Position{16, 2},
				BottomRight: Position{18, 4},
			}
		}
		return Area{
			TopLeft:     Position{16, 17},
			BottomRight: Position{18, 19},
		}
	default:
		return Area{
			TopLeft:     Position{0, 0},
			BottomRight: Position{0, 0},
		}
	}
}

// ---------- Core Logic ----------

func (t *Tower) Clone(mode string, level int) *Tower {
	maxHP := t.MaxHP
	if mode == "enhanced" {
		maxHP *= 1 + 0.1*float64(level)
	}
	return &Tower{
		Type:        t.Type,
		MaxHP:       maxHP,
		HP:          maxHP,
		ATK:         t.ATK,
		DEF:         t.DEF,
		CRIT:        t.CRIT,
		EXP:         t.EXP,
		Range:       t.Range,
		AttackSpeed: t.AttackSpeed,
	}
}

func (t *Tower) TakeDamage(rawAtk float64, attackerLevel int) (float64, bool) {
	dmg := rawAtk - t.DEF/1.5
	if dmg < 0 {
		dmg = 0
	}
	t.HP -= dmg
	if t.HP < 0 {
		t.HP = 0
	}
	return dmg, t.HP == 0
}

func (t *Tower) Heal(amount float64) {
	t.HP += amount
	if t.HP > t.MaxHP {
		t.HP = t.MaxHP
	}
}

// ---------- Utilities ----------

func (a Area) Contains(pos Position) bool {
	return pos.X >= a.TopLeft.X && pos.X <= a.BottomRight.X &&
		pos.Y >= a.TopLeft.Y && pos.Y <= a.BottomRight.Y
}

func GetLowestHPTower(player *Player) *Tower {
	var lowest *Tower
	for _, tower := range player.Towers {
		if tower.HP <= 0 {
			continue
		}
		if lowest == nil || tower.HP < lowest.HP {
			lowest = tower
		}
	}
	return lowest
}
