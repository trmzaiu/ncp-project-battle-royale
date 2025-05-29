package game

import (
	"fmt"
	"royaka/internal/model"
)

// ===================== Public Turn APIs =====================

// PlayTurnSimple handles attacking a tower using a troop.
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

	message := fmt.Sprintf("%s dealt %f damage to %s", troop.Name, damage, targetTower.Type)
	if isCrit {
		message += " (Critical hit!)"
	}
	if destroyed {
		message += " and destroyed it!"
		player.Mana += 3
		if player.Mana > 10 {
			player.Mana = 10
		}
		// g.TurnTimerCancel()
		g.StartTurnTimer()
	} else {
		// g.TurnTimerCancel()
		g.SwitchTurn()
	}

	player.Turn++
	return int(damage), isCrit, message
}

// HealTower allows a healing troop to heal the lowest-HP tower.
func (g *Game) HealTower(player *model.Player, troop *model.Troop) (int, *model.Tower, string) {
	if player.Mana < troop.MANA {
		return 0, nil, manaRequestMessage
	}
	if troop.Type != "healer" {
		return 0, nil, "Only healing troop can heal towers"
	}
	player.Mana -= troop.MANA

	lowest := model.GetLowestHPTower(player)
	if lowest == nil {
		return 0, nil, "No tower found to heal"
	}

	healAmount, isCrit := troop.CalculateHeal(player.User.Level)
	lowest.HP += float64(healAmount)
	if lowest.HP > lowest.MaxHP {
		lowest.HP = lowest.MaxHP
	}

	message := fmt.Sprintf("%s healed %s tower for %f HP", troop.Name, lowest.Type, healAmount)
	if isCrit {
		message += " (Critical heal!)"
	}

	player.Turn++
	// g.TurnTimerCancel()
	g.SwitchTurn()

	return int(healAmount), lowest, message
}

// ===================== Combat Core =====================

// AttackTower applies damage from troop to tower and returns result.
func (g *Game) AttackTower(player *model.Player, troop *model.Troop, tower *model.Tower) (float64, bool, bool) {
	atk, isCrit := troop.CalculateDamage(player.User.Level)
	damageDealt, destroyed := tower.TakeDamage(atk, player.User.Level)
	return damageDealt, isCrit, destroyed
}

// getTargetTower returns the opponent's tower based on target string.
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
