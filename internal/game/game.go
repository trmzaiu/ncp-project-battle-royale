// internal/game/game.go

package game

import (
	"royaka/internal/model"
	"royaka/internal/utils"
)

type Game struct {
	Player1 *model.Player
	Player2 *model.Player
	Turn    int // 1 or 2
	Started bool
}

// NewGame initializes a new game with two players.
func NewGame(p1, p2 *model.Player) *Game {
	return &Game{
		Player1: p1,
		Player2: p2,
		Turn:    1,
		Started: true,
	}
}

func (g *Game) CurrentPlayer() *model.Player {
	if g.Turn == 1 {
		return g.Player1
	}
	return g.Player2
}

func (g *Game) Opponent() *model.Player {
	if g.Turn == 1 {
		return g.Player2
	}
	return g.Player1
}

// AttackTroop simulates an attack on a tower by a troop.
func (g *Game) AttackTroop(troop model.Troop, target *model.Tower, critEnabled bool) (damage int, isDestroyed bool) {
	crit := false
	atk := troop.ATK

	if critEnabled {
		crit = utils.IsCriticalHit(int(troop.CRIT))
		if crit {
			atk = int(float64(atk) * 1.2)
		}
	}

	rawDmg := atk - target.DEF
	if rawDmg < 0 {
		rawDmg = 0
	}

	target.HP -= rawDmg
	if target.HP <= 0 {
		target.HP = 0
	}

	return rawDmg, target.HP == 0
}

// Play a turn
func (g *Game) PlayTurn(troop model.Troop, towerType string, critEnabled bool) (result string) {
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

	result = troop.Name + " dealt " + utils.Itoa(damage) + " damage to " + towerType
	if destroyed {
		result += " and destroyed it!"
	}

	g.Turn = 3 - g.Turn
	return result
}

// Check winner
func (g *Game) CheckWinner() string {
	if g.Player1.Towers["king"].HP <= 0 {
		return g.Player2.Username + " wins!"
	}
	if g.Player2.Towers["king"].HP <= 0 {
		return g.Player1.Username + " wins!"
	}
	return ""
}

// Reset game state
func (g *Game) Reset() {
	g.Player1.Reset()
	g.Player2.Reset()
	g.Turn = 1
	g.Started = false
}
