// internal/game/game.go

package game

import (
	"fmt"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"
)

type Game struct {
	Player1   *model.Player
	Player2   *model.Player
	Turn      int // For simple mode (1 or 2)
	Started   bool
	Enhanced  bool
	StartTime time.Time
	MaxTime   time.Duration // For enhanced mode
}

// NewGame initializes a new game with two players.
func NewGame(p1, p2 *model.Player, enhanced bool) *Game {
	game := &Game{
		Player1:  p1,
		Player2:  p2,
		Turn:     2,
		Started:  true,
		Enhanced: enhanced,
	}

	if enhanced {
		game.StartTime = time.Now()
		game.MaxTime = 3 * time.Minute
	}

	return game
}

func (g *Game) CurrentPlayer() *model.Player {
	if g.Enhanced {
		return nil // No turn-based play in enhanced mode
	}
	if g.Turn == 1 {
		return g.Player1
	}
	return g.Player2
}

func (g *Game) Opponent(p *model.Player) *model.Player {
	if g.Player1.User.Username == p.User.Username {
		return g.Player2
	}
	return g.Player1
}

// AttackTower simulates an attack on a tower by a troop.
func (g *Game) AttackTower(p *model.Player, troop *model.Troop, target *model.Tower, critEnabled bool) (int, bool, string) {
	atk := troop.CalculateDamage(p.User.Level, critEnabled)
	dmgToTower, destroyed := target.TakeDamage(atk, p.User.Level)

	if target.HP > 0 {
		counterDmg := target.CounterDamage()
		troop.HP -= counterDmg
		if troop.HP < 0 {
			troop.HP = 0
		}
		log := fmt.Sprintf("Tower retaliated for %d damage.", counterDmg)
		if troop.HP == 0 {
			log += " Troop was defeated!"
		}
		return dmgToTower, destroyed, log
	}

	return dmgToTower, destroyed, ""
}

// PlayTurn - handles one attack turn (Simple Mode only)
func (g *Game) PlayTurn(p *model.Player, troop *model.Troop, towerType string) string {
	if g.Enhanced && time.Since(g.StartTime) > g.MaxTime {
		return "Time is up!"
	}

	if g.Enhanced {
		return g.ApplySpecialSkill(p, troop)
	}

	if p.Mana < troop.MANA {
		return "Not enough mana!"
	}
	p.Mana -= troop.MANA

	target, err := g.getTargetTower(p, towerType)
	if err != nil {
		return err.Error()
	}

	if troop.Name == "Queen" {
		return g.handleQueenAction(p)
	}

	return g.handleAttackAction(p, troop, target, towerType)
}

func (g *Game) getTargetTower(p *model.Player, towerType string) (*model.Tower, error) {
	op := g.Opponent(p)
	switch towerType {
	case "guard1":
		return op.Towers["guard1"], nil
	case "guard2":
		return op.Towers["guard2"], nil
	case "king":
		return op.Towers["king"], nil
	default:
		return nil, fmt.Errorf("invalid tower")
	}
}

func (g *Game) handleQueenAction(p *model.Player) string {
	var lowest *model.Tower
	for _, tower := range p.Towers {
		if lowest == nil || tower.HP < lowest.HP {
			lowest = tower
		}
	}
	if lowest != nil {
		lowest.Heal(300)
		result := "Queen healed " + lowest.Type + " tower by 300 HP!"
		if !g.Enhanced {
			g.Turn = 3 - g.Turn
		}
		return result
	}
	return "Queen could not find a tower to heal."
}

func (g *Game) handleAttackAction(p *model.Player, troop *model.Troop, target *model.Tower, towerType string) string {
	dmg, destroyed, _ := g.AttackTower(p, troop, target, g.Enhanced)
	result := troop.Name + " dealt " + utils.Itoa(dmg) + " damage to " + towerType
	if destroyed {
		result += " and destroyed it!"
	}
	return result
}

func (g *Game) ApplySpecialSkill(p *model.Player, t *model.Troop) string {
	if !g.Enhanced || time.Since(g.StartTime) > g.MaxTime {
		return "Special skills are only available in Enhanced mode."
	}

	switch t.Special {
	case "Shield":
		// Apply shield to all towers
		p.ApplyDefenseBoost(0.2)
		return "Shield applied! Defense increased for all towers."
	case "Attack Boost":
		// Apply attack boost to all troops
		p.BoostAllTroops()
		return "Attack Boost applied! Damage increased for all troops."
	case "Fortify":
		// Apply fortify to all troops
		t.FortifyHP(50)
		return "Fortify applied! Troop's HP increased."
	case "Double Strike":
		// Attack twice
		var target *model.Tower
		opponentTowers := g.Opponent(p).Towers

		if opponentTowers["guard1"].HP > 0 {
			target = opponentTowers["guard1"]
		} else if opponentTowers["guard2"].HP > 0 {
			target = opponentTowers["guard2"]
		} else {
			target = opponentTowers["king"]
		}
		g.AttackTower(p, t, target, false)
		g.AttackTower(p, t, target, false)
		return "Double Strike applied! Troop attacks " + target.Type + " twice!"
	case "Charge":
		// Charge mana
		p.FullyChargeMana()
		return "Charge applied! Mana fully restored."
	case "Heal":
		var lowest *model.Tower
		for _, tower := range p.Towers {
			if lowest == nil || tower.HP < lowest.HP {
				lowest = tower
			}
		}
		if lowest != nil {
			lowest.Heal(300)
			return "Heal applied! " + lowest.Type + " tower HP restored."
		}
	}
	return "Invalid special skill."
}

// CheckWinner returns result string if winner is found
func (g *Game) CheckWinner() string {
	if g.Player1.Towers["king"].HP <= 0 {
		AwardEXP(g.Player2.User, g.Player1.User, false)
		return g.Player2.User.Username + " wins!"
	}
	if g.Player2.Towers["king"].HP <= 0 {
		AwardEXP(g.Player1.User, g.Player2.User, false)
		return g.Player1.User.Username + " wins!"
	}
	if g.Enhanced && time.Since(g.StartTime) > g.MaxTime {
		// Compare destroyed towers
		p1Score := g.Player1.DestroyedCount()
		p2Score := g.Player2.DestroyedCount()
		if p1Score > p2Score {
			AwardEXP(g.Player1.User, g.Player2.User, false)
			return g.Player1.User.Username + " wins by score!"
		}
		if p2Score > p1Score {
			AwardEXP(g.Player2.User, g.Player1.User, false)
			return g.Player2.User.Username + " wins by score!"
		}
		AwardEXP(g.Player1.User, g.Player2.User, true)
		return "It's a draw!"
	}
	return ""
}

// AwardEXP updates the user's EXP, level, and match records
func AwardEXP(winner, loser *model.User, isDraw bool) {
	winner.GamesPlayed++
	loser.GamesPlayed++

	if isDraw {
		winner.AddExp(10)
		loser.AddExp(10)
	} else {
		winner.GamesWon++
		winner.AddExp(30)
	}

	model.SaveUser(*winner)
	model.SaveUser(*loser)
}

func (g *Game) Reset() {
	g.Player1.Reset()
	g.Player2.Reset()
	g.Turn = 1
	g.Started = false
}
