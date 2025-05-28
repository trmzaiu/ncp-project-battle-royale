package game

import (
	"fmt"
	"math"
	"royaka/internal/model"
	"time"
)

// =============================================================================
// 1. HỆ THỐNG TÌM KIẾM MỤC TIÊU
// =============================================================================

// findClosestEnemyTroopForTower - Tìm troop địch gần nhất cho tower tấn công
func (g *Game) findClosestEnemyTroop(tower *model.TowerInstance) *model.TroopInstance {
	if tower == nil || tower.Template == nil {
		return nil
	}

	var closestTroop *model.TroopInstance
	minDist := math.MaxFloat64

	// Tính vị trí center của tower
	towerPos := tower.GetPosition()

	// Duyệt qua tất cả troop trong BattleMap
	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			if troop, ok := entity.(*model.TroopInstance); ok {
				// Chỉ tấn công troop địch còn sống
				if troop.Owner != tower.Owner && troop.IsAlive() {
					dist := calculateDistance(towerPos, troop.Position)
					// Kiểm tra trong tầm và gần hơn target hiện tại
					if dist <= tower.Template.Range && dist < minDist {
						closestTroop = troop
						minDist = dist
					}
				}
			}
		}
	}

	return closestTroop
}

// =============================================================================
// 2. HỆ THỐNG PHÒNG THỦ
// =============================================================================

// towerAttackTroop - Xử lý tower tấn công troop
func (g *Game) towerAttackTroop(tower *model.TowerInstance, target *model.TroopInstance) {
	if tower == nil || target == nil || tower.Template == nil || target.Template == nil {
		return
	}

	currentTime := time.Now()
	// Tính cooldown tấn công
	attackCooldown := time.Duration(tower.Template.AttackSpeed * float64(time.Second))

	// Kiểm tra cooldown
	if currentTime.Sub(tower.LastAttackTime) < attackCooldown {
		return
	}

	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	if !target.IsAlive() {
		return
	}

	// Gây damage lên troop
	damage := math.Max(tower.Template.ATK, 1)
	target.Template.HP -= damage
	tower.LastAttackTime = currentTime

	fmt.Printf("Tower %s attacks troop %s for %.1f damage. Troop HP: %.1f\n",
		tower.Template.Type, target.Template.Name, damage, target.Template.HP)

	// Kiểm tra troop có chết không
	if target.Template.HP <= 0 {
		target.IsDead = true

		g.addKillReward(tower.Owner, target)
		fmt.Printf("Troop %s killed by tower!\n", target.Template.Name)
	}
}