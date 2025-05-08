// internal/game/mechanics.go

package game

import "royaka/internal/player"

func AttackTower(troop *player.Troop, target *player.Tower) int {
	dmg := troop.DamageTo(target)
	target.TakeDamage(dmg)
	return dmg
}

func HealLowestTower(towers map[string]*player.Tower) {
	var lowest *player.Tower
	for _, t := range towers {
		if t.HP > 0 && (lowest == nil || t.HP < lowest.HP) {
			lowest = t
		}
	}
	if lowest != nil {
		lowest.HP += 300
	}
}
