// internal/game/game.go

package game

import (
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"
)

type Game struct {
	Player1 *model.Player
	Player2 *model.Player
	Turn    int // 1 or 2, only used in Simple Mode
	Started bool
	StartAt time.Time
	Mode    string // "simple" or "enhanced"
}

// NewGame initializes a new game with two players.
func NewGame(p1, p2 *model.Player, mode string) *Game {
	return &Game{
		Player1: p1,
		Player2: p2,
		Turn:    1,
		Started: true,
		StartAt: time.Now(),
		Mode:    mode,
	}
}

func (g *Game) CurrentPlayer() *model.Player {
	if g.Mode == "enhanced" {
		return nil // Not turn-based
	}
	if g.Turn == 1 {
		return g.Player1
	}
	return g.Player2
}

func (g *Game) Opponent() *model.Player {
	if g.Mode == "enhanced" {
		return nil // Not turn-based
	}
	if g.Turn == 1 {
		return g.Player2
	}
	return g.Player1
}

// AttackTroop simulates an attack on a tower by a troop.
func (g *Game) AttackTroop(troop model.Troop, target *model.Tower, critEnabled bool) (damage int, isDestroyed bool) {
	atk := troop.ATK

	if critEnabled {
		if utils.IsCriticalHit(int(troop.CRIT)) {
			atk = int(float64(atk) * 1.2)
		}
	}

	rawDmg := atk - target.DEF
	if rawDmg < 0 {
		rawDmg = 0
	}
	target.HP -= rawDmg
	if target.HP < 0 {
		target.HP = 0
	}

	return rawDmg, target.HP == 0
}

// PlayTurn - handles one attack turn (Simple Mode only)
func (g *Game) PlayTurn(troop model.Troop, towerType string, critEnabled bool) string {
	if g.Mode != "simple" {
		return "PlayTurn only valid in Simple Mode"
	}

	opponent := g.Opponent()
	var target *model.Tower

	switch towerType {
	case "guard1":
		target = opponent.Towers["guard1"]
	case "guard2":
		if opponent.Towers["guard1"].HP > 0 {
			return "You must destroy Guard Tower 1 first!"
		}
		target = opponent.Towers["guard2"]
	case "king":
		if opponent.Towers["guard1"].HP > 0 || opponent.Towers["guard2"].HP > 0 {
			return "Destroy both Guard Towers before attacking King Tower!"
		}
		target = opponent.Towers["king"]
	default:
		return "Invalid target tower."
	}

	damage, destroyed := g.AttackTroop(troop, target, critEnabled)
	result := troop.Name + " dealt " + utils.Itoa(damage) + " damage to " + towerType
	if destroyed {
		result += " and destroyed it!"
	}

	g.Turn = 3 - g.Turn
	return result
}

// CheckWinner returns result string if winner is found
func (g *Game) CheckWinner() string {
	if g.Player1.Towers["king"].HP <= 0 {
		return g.Player2.Username + " wins!"
	}
	if g.Player2.Towers["king"].HP <= 0 {
		return g.Player1.Username + " wins!"
	}
	return ""
}

// GetWinnerAfterTime returns the winner based on number of towers left (Enhanced Mode)
func (g *Game) GetWinnerAfterTime() string {
	if g.Mode != "enhanced" {
		return ""
	}
	p1Towers := g.countRemainingTowers(g.Player1)
	p2Towers := g.countRemainingTowers(g.Player2)

	if p1Towers > p2Towers {
		return g.Player1.Username + " wins!"
	} else if p2Towers > p1Towers {
		return g.Player2.Username + " wins!"
	} else {
		return "Draw"
	}
}

func (g *Game) countRemainingTowers(p *model.Player) int {
	total := 0
	for _, t := range p.Towers {
		if t.HP > 0 {
			total++
		}
	}
	return total
}

// Reset game state
func (g *Game) Reset() {
	g.Player1.Reset()
	g.Player2.Reset()
	g.Turn = 1
	g.Started = false
	g.StartAt = time.Time{}
}