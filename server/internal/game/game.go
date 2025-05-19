package game

import (
	"fmt"
	"math/rand"
	"royaka/internal/model"
	"time"
)

type Game struct {
	Player1   *model.Player
	Player2   *model.Player
	Turn      string
	Started   bool
	Enhanced  bool
	StartTime time.Time
	MaxTime   time.Duration
	TickRate  float64
	LastTick  time.Time
}

// ===================== Game Initialization =====================

func NewGame(p1, p2 *model.Player, enhanced bool) *Game {
	startingPlayer := p1.User.Username
	if rand.Intn(2) == 0 {
		startingPlayer = p2.User.Username
	}

	game := &Game{
		Player1:  p1,
		Player2:  p2,
		Turn:     startingPlayer,
		Started:  true,
		Enhanced: enhanced,
	}

	if enhanced {
		game.StartTime = time.Now()
		game.MaxTime = 3 * time.Minute
	}

	return game
}

// ===================== Turn Management =====================

func (g *Game) CurrentPlayer() *model.Player {
	if g.Enhanced {
		return nil // No turn-based play in enhanced mode
	}
	if g.Turn == g.Player1.User.Username {
		return g.Player1
	}
	return g.Player2
}

func (g *Game) SwitchTurn() {
	if g.Turn == g.Player1.User.Username {
		g.Turn = g.Player2.User.Username
	} else {
		g.Turn = g.Player1.User.Username
	}

	nextPlayer := g.CurrentPlayer()
	if nextPlayer.Turn > 0 {
		nextPlayer.Mana += 3
		if nextPlayer.Mana > 10 {
			nextPlayer.Mana = 10
		}
	}
}

func (g *Game) SkipTurn(player *model.Player) {
	player.Turn++
	g.SwitchTurn()
}

// ===================== Turn Actions =====================

func (g *Game) PlayTurnSimple(player *model.Player, troop *model.Troop, tower string) (int, bool, string) {
	if player.Mana < troop.MANA {
		return 0, false, manaRequestMessage
	}
	player.Mana -= troop.MANA

	if tower == "king" {
		op := g.Opponent(player)
		if op.Towers["guard1"].HP > 0 || op.Towers["guard2"].HP > 0 {
			player.Mana += troop.MANA
			return 0, false, "You must destroy both guard towers before attacking the king!"
		}
	}

	targetTower, err := g.getTargetTower(player, tower)
	if err != nil {
		return 0, false, "Invalid tower target"
	}

	damage, isCrit, destroyed := g.AttackTower(player, troop, targetTower)

	message := fmt.Sprintf("%s dealt %d damage to %s", troop.Name, damage, targetTower.Type)
	if isCrit {
		message += " (Critical hit!)"
	}
	if destroyed {
		message += " and destroyed it!"
	}

	player.Turn++

	if !destroyed {
		g.SwitchTurn()
	} else {
		player.Mana += 3
		if player.Mana > 10 {
			player.Mana = 10
		}
	}

	return damage, isCrit, message
}

func (g *Game) PlayTurnEnhanced(player *model.Player, troop *model.Troop) {
	return // not implemented yet
}

func (g *Game) HealTower(player *model.Player, troop *model.Troop) (int, *model.Tower, string) {
	if player.Mana < troop.MANA {
		return 0, nil, manaRequestMessage
	}
	player.Mana -= troop.MANA

	if troop.Type != "heal" {
		return 0, nil, "Only healing troop can heal towers"
	}

	lowest := model.GetLowestHPTower(player)
	if lowest == nil {
		return 0, nil, "No tower found to heal"
	}

	healAmount, isCrit := troop.CalculateHeal(player.User.Level)

	lowest.HP += healAmount
	if lowest.HP > lowest.MaxHP {
		lowest.HP = lowest.MaxHP
	}

	message := fmt.Sprintf("Queen healed %s tower for %d HP", lowest.Type, healAmount)
	if isCrit {
		message += " (Critical heal!)"
	}

	player.Turn++
	g.SwitchTurn()

	return healAmount, lowest, message
}

// ===================== Combat =====================

func (g *Game) AttackTower(player *model.Player, troop *model.Troop, tower *model.Tower) (int, bool, bool) {
	atk, isCrit := troop.CalculateDamage(player.User.Level)
	damageDealt, destroyed := tower.TakeDamage(atk, player.User.Level)
	return damageDealt, isCrit, destroyed
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

// ===================== Game Outcome =====================

func (g *Game) CheckWinner() (*model.Player, string) {
	if g.Player1.Towers["king"].HP <= 0 {
		g.Started = false
		if !g.Started {
			AwardEXP(g.Player2.User, g.Player1.User, false)
		}
		return g.Player2, g.Player2.User.Username + " wins!"
	}

	if g.Player2.Towers["king"].HP <= 0 {
		g.Started = false
		if !g.Started {
			AwardEXP(g.Player1.User, g.Player2.User, false)
		}
		return g.Player1, g.Player1.User.Username + " wins!"
	}

	if g.Enhanced && time.Since(g.StartTime) > g.MaxTime {
		p1Score := g.Player1.DestroyedCount()
		p2Score := g.Player2.DestroyedCount()

		if p1Score > p2Score {
			AwardEXP(g.Player1.User, g.Player2.User, false)
			return g.Player1, g.Player1.User.Username + " wins by score!"
		}
		if p2Score > p1Score {
			AwardEXP(g.Player2.User, g.Player1.User, false)
			return g.Player2, g.Player2.User.Username + " wins by score!"
		}

		g.Started = false
		if !g.Started {
			AwardEXP(g.Player1.User, g.Player2.User, true)
		}
		return nil, "It's a draw!"
	}

	return nil, ""
}

func (g *Game) SetWinner(winner *model.Player) {
	g.Started = false
	if winner == g.Player1 {
		AwardEXP(g.Player1.User, g.Player2.User, false)
	} else if winner == g.Player2 {
		AwardEXP(g.Player2.User, g.Player1.User, false)
	}
}

func AwardEXP(winner, loser *model.User, isDraw bool) {
	if isDraw {
		winner.AddExp(10)
		loser.AddExp(10)
	} else {
		winner.GamesWon++
		winner.AddExp(30)
	}

	winner.GamesPlayed++
	loser.GamesPlayed++

	model.SaveUser(winner)
	model.SaveUser(loser)
}

// ===================== Game Utility =====================

func (g *Game) Opponent(p *model.Player) *model.Player {
	if g.Player1.User.Username == p.User.Username {
		return g.Player2
	}
	return g.Player1
}

func (g *Game) Reset(mode string) {
	g.Player1.Reset(mode)
	g.Player2.Reset(mode)
	g.Turn = ""
	g.Enhanced = false
	g.StartTime = time.Time{}
	g.Started = false
}
