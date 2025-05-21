// internal/model/tower.go

package model

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
)

type Tower struct {
	Type  string  `json:"type"`
	MaxHP float64 `json:"max_hp"`
	HP    float64 `json:"hp"`
	ATK   float64 `json:"atk"`
	DEF   float64 `json:"def"`
	CRIT  float64 `json:"crit"`
	EXP   int     `json:"exp"`
}

type Area struct {
	TopLeft     Position `json:"top_left"`
	BottomRight Position `json:"bottom_right"`
}

type TowerInstance struct {
	ID          string `json:"id"`
	Template    *Tower `json:"template"`
	TypeEntity  string `json:"type_entity"`
	Owner       string `json:"owner"`
	Area        Area   `json:"area"`
	IsDestroyed bool   `json:"is_destroyed"`
}

func (t *TowerInstance) GetID() string {
	return t.ID
}
func (t *TowerInstance) GetOwner() string {
	return t.Owner
}

func (t *TowerInstance) GetType() string {
	return t.TypeEntity
}

// For position of Tower, get center or top-left (depends on your logic)
func (t *TowerInstance) GetPosition() Position {
	return Position{
		X: (t.Area.TopLeft.X + t.Area.BottomRight.X) / 2,
		Y: (t.Area.TopLeft.Y + t.Area.BottomRight.Y) / 2,
	}
}
func (t *TowerInstance) IsAlive() bool {
	return !t.IsDestroyed
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

func (a Area) Contains(pos Position) bool {
	return pos.X >= a.TopLeft.X && pos.X <= a.BottomRight.X &&
		pos.Y >= a.TopLeft.Y && pos.Y <= a.BottomRight.Y
}

func (t *Tower) Clone() *Tower {
	return &Tower{
		Type:  t.Type,
		MaxHP: t.MaxHP,
		HP:    t.MaxHP,
		ATK:   t.ATK,
		DEF:   t.DEF,
		CRIT:  t.CRIT,
		EXP:   t.EXP,
	}
}

func (t *Tower) TakeDamage(rawAtk float64, attackerLevel int) (float64, bool) {
	dmg := rawAtk - t.DEF/2
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
	t.DEF = float64(t.DEF) * (1 + percent)
}

func (t *Tower) Heal(amount float64) {
	t.HP += amount
	if t.HP > t.MaxHP {
		t.HP = t.MaxHP
	}
}

func (t *Tower) CounterDamage() float64 {
	rand.Seed(time.Now().UnixNano())

	baseDamage := t.ATK

	if rand.Float64() < t.CRIT {
		baseDamage = float64(baseDamage) * 1.5
	}

	return baseDamage
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

func (t *Tower) Reset(key string) {
	if def, ok := defaultTowers[key]; ok {
		t.MaxHP = def.MaxHP
		t.HP = def.MaxHP
		t.ATK = def.ATK
		t.DEF = def.DEF
		t.CRIT = def.CRIT
		t.EXP = def.EXP
	}
}

func CreateTowerInstances(towers map[string]*Tower, owner string, isPlayer1 bool) []*TowerInstance {
	instances := []*TowerInstance{}
	for _, tower := range towers {
		instance := &TowerInstance{
			ID:          uuid.New().String(),
			Template:    tower,
			TypeEntity:  "tower",
			Owner:       owner,
			IsDestroyed: false,
			Area:        GetTowerArea(tower.Type, isPlayer1),
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
				TopLeft:     Position{X: 9, Y: 0},
				BottomRight: Position{X: 11, Y: 2},
			}
		}
		return Area{
			TopLeft:     Position{X: 9, Y: 18},
			BottomRight: Position{X: 11, Y: 20},
		}
	case "guard1":
		if isPlayer1 {
			return Area{
				TopLeft:     Position{X: 4, Y: 2},
				BottomRight: Position{X: 5, Y: 3},
			}
		}
		return Area{
			TopLeft:     Position{X: 4, Y: 17},
			BottomRight: Position{X: 5, Y: 18},
		}
	case "guard2":
		if isPlayer1 {
			return Area{
				TopLeft:     Position{X: 15, Y: 2},
				BottomRight: Position{X: 16, Y: 3},
			}
		}
		return Area{
			TopLeft:     Position{X: 15, Y: 17},
			BottomRight: Position{X: 16, Y: 18},
		}
	default:
		return Area{
			TopLeft:     Position{X: 0, Y: 0},
			BottomRight: Position{X: 0, Y: 0},
		}
	}
}
