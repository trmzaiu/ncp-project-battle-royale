// internal/game/game.go

package game

import (
	"royaka/internal/player"
	"time"
)

type BattleState struct {
	StartTime  time.Time
	Duration   time.Duration
	Player1    *player.Player
	Player2    *player.Player
	Winner     *player.Player
	IsFinished bool
}

func NewBattle(p1, p2 *player.Player) *BattleState {
	return &BattleState{
		StartTime:  time.Now(),
		Duration:   3 * time.Minute,
		Player1:    p1,
		Player2:    p2,
		IsFinished: false,
	}
}

func (b *BattleState) CheckEndCondition() {
	if time.Since(b.StartTime) >= b.Duration {
		// Calculate winner based on tower HP
		t1 := b.Player1.Towers["king"].HP + b.Player1.Towers["guard1"].HP + b.Player1.Towers["guard2"].HP
		t2 := b.Player2.Towers["king"].HP + b.Player2.Towers["guard1"].HP + b.Player2.Towers["guard2"].HP
		if t1 > t2 {
			b.Winner = b.Player1
		} else if t2 > t1 {
			b.Winner = b.Player2
		} // else draw (nil)
		b.IsFinished = true
	}
}

func (b *BattleState) DeployTroop(player *player.Player, troop *player.Troop, lane string) bool {
	if player.Mana >= troop.MANA {
		player.Mana -= troop.MANA
		player.Troops = append(player.Troops, troop)
		return true
	}
	return false
}
