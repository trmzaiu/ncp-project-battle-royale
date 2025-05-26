package game

import (
	"fmt"
	"math"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"
)

// =============================================================================
// 1. HỆ THỐNG TÌM KIẾM MỤC TIÊU
// =============================================================================

// findLowestHPAllyInRange - Tìm troop có HP thấp nhất trong tầm nhìn của healer (CẢI THIỆN)
func (g *Game) findLowestHPAllyInRange(healer *model.TroopInstance) *model.TroopInstance {
	if healer == nil || healer.Template == nil {
		return nil
	}

	var lowestHPAlly *model.TroopInstance
	minHPPercent := 0.9 // Chỉ heal khi HP < 90%
	healerPos := healer.Position
	healRange := healer.Template.Range

	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			if ally, ok := entity.(*model.TroopInstance); ok &&
				ally.IsAlive() &&
				ally.Owner == healer.Owner &&
				ally.ID != healer.ID {

				dist := calculateDistance(healerPos, ally.Position)

				if dist <= healRange {
					hpPercent := ally.Template.HP / ally.Template.MaxHP

					// Ưu tiên heal ally có HP thấp nhất và dưới ngưỡng
					if hpPercent < minHPPercent {
						minHPPercent = hpPercent
						lowestHPAlly = ally
					}
				}
			}
		}
	}

	return lowestHPAlly
}

// findAllyInRange - Tìm bất kỳ đồng minh nào trong phạm vi cho trước
func (g *Game) findAllyInRange(healer *model.TroopInstance, searchRange float64) *model.TroopInstance {
	if healer == nil || healer.Template == nil {
		return nil
	}

	healerPos := healer.Position

	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			if ally, ok := entity.(*model.TroopInstance); ok &&
				ally.IsAlive() &&
				ally.Owner == healer.Owner &&
				ally.ID != healer.ID {

				dist := calculateDistance(healerPos, ally.Position)
				if dist <= searchRange {
					return ally // Tìm thấy đồng minh trong phạm vi
				}
			}
		}
	}

	return nil // Không có đồng minh nào trong phạm vi
}

// findAllyToFollow - Tìm ally tốt nhất để healer đi theo
func (g *Game) findAllyToFollow(healer *model.TroopInstance) *model.TroopInstance {
	if healer == nil || healer.Template == nil {
		return nil
	}

	var bestAlly *model.TroopInstance
	bestScore := 0.0
	healerPos := healer.Position
	isPlayer1 := healer.Owner == g.Player1.User.Username

	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			ally, ok := entity.(*model.TroopInstance)
			if !ok || !g.isValidAllyToFollow(healer, ally) {
				continue
			}
			score := g.calculateAllyFollowScore(healerPos, ally, isPlayer1)
			if score > bestScore {
				bestScore = score
				bestAlly = ally
			}
		}
	}

	return bestAlly
}

// isValidAllyToFollow checks if the ally is a valid candidate to follow.
func (g *Game) isValidAllyToFollow(healer, ally *model.TroopInstance) bool {
	return ally != nil &&
		ally.IsAlive() &&
		ally.Owner == healer.Owner &&
		ally.ID != healer.ID &&
		ally.Template.Type != "healer"
}

// calculateAllyFollowScore calculates the score for an ally to be followed.
func (g *Game) calculateAllyFollowScore(healerPos model.Position, ally *model.TroopInstance, isPlayer1 bool) float64 {
	dist := calculateDistance(healerPos, ally.Position)
	hpPercent := ally.Template.HP / ally.Template.MaxHP

	score := 0.0

	// Ưu tiên ally gần
	if dist <= 8 {
		score += (8 - dist)
	}

	// Ưu tiên ally khỏe mạnh
	score += hpPercent * 3

	// Ưu tiên ally không ở quá sâu trong phe địch
	allyInEnemyTerritory := g.isHealerInEnemyTerritory(&model.TroopInstance{
		Position: ally.Position,
	}, isPlayer1)

	if !allyInEnemyTerritory {
		score += 5
	}

	// Ưu tiên damage dealer
	if ally.Template.DMG > ally.Template.HP/5 {
		score += 2
	}

	return score
}

// isHealerInEnemyTerritory - Kiểm tra healer có đang ở phe địch không
func (g *Game) isHealerInEnemyTerritory(healer *model.TroopInstance, isPlayer1 bool) bool {
	if isPlayer1 {
		// Player 1 spawn từ dưới (Y=0), phe địch là Y > 15
		return healer.Position.Y > 14.0
	} else {
		// Player 2 spawn từ trên (Y=21), phe địch là Y < 6
		return healer.Position.Y < 7.0
	}
}

// =============================================================================
// 2. HÀNH VI DI CHUYỂN CỦA HEALER
// =============================================================================

// searchForAlliesSlowly - Tìm kiếm đồng minh một cách chậm rãi
func (g *Game) searchForAlliesSlowly(healer *model.TroopInstance, speed float64, isPlayer1 bool) {
	// Di chuyển chậm về trung tâm theo trục X
	centerX := MAP_SIZE / 2

	if healer.Position.X < centerX-2 {
		healer.Position.X += speed * 0.3
	} else if healer.Position.X > centerX+2 {
		healer.Position.X -= speed * 0.3
	}

	// Di chuyển rất chậm về phía trước để không bỏ lỡ đồng minh
	directionY := getDirectionY(isPlayer1)
	healer.Position.Y += directionY * speed * 0.2

	// Không tiến quá xa khỏi vùng spawn
	var maxAdvanceY float64
	if isPlayer1 {
		maxAdvanceY = 9.0 // Player 1 không tiến quá Y = 10
	} else {
		maxAdvanceY = 12.0 // Player 2 không lùi quá Y = 11
	}

	if (isPlayer1 && healer.Position.Y > maxAdvanceY) ||
		(!isPlayer1 && healer.Position.Y < maxAdvanceY) {
		healer.Position.Y = maxAdvanceY
	}
}

// followAlly - Follow đồng minh với khoảng cách an toàn
func (g *Game) followAlly(healer *model.TroopInstance, ally *model.TroopInstance, speed float64) {
	if healer == nil || ally == nil {
		return
	}

	idealDistance := healer.Template.Range  // Khoảng cách lý tưởng để follow (không quá gần, không quá xa)
	currentDist := calculateDistance(healer.Position, ally.Position)

	if currentDist > idealDistance {
		// Quá xa -> di chuyển lại gần
		g.moveTowardPosition(healer, ally.Position, speed*0.9)
	} else if currentDist < idealDistance {
		// Quá gần -> lùi lại một chút để tránh chen chúc
		g.moveAwayFromPosition(healer, ally.Position, speed*0.5)
	} else {
		// Khoảng cách vừa phải -> di chuyển cùng hướng với ally
		g.moveInSameDirection(healer, ally, speed*0.7)
	}
}

// handleHealerWithoutAllies - Xử lý healer khi không có đồng minh
func (g *Game) handleHealerWithoutAllies(healer *model.TroopInstance, speed float64, isPlayer1 bool) {
	// Kiểm tra vị trí hiện tại
	if g.isHealerInEnemyTerritory(healer, isPlayer1) {
		// Ở phe địch -> quay về
		g.moveHealerBackToSafety(healer, speed, isPlayer1)
	} else {
		// Ở phe mình -> di chuyển chậm để tìm hoặc chờ đồng minh
		g.searchForAlliesSlowly(healer, speed, isPlayer1)
	}
}

// moveTowardPosition - Di chuyển đến vị trí mục tiêu
func (g *Game) moveTowardPosition(troop *model.TroopInstance, targetPos model.Position, speed float64) {
	dirX := targetPos.X - troop.Position.X
	dirY := targetPos.Y - troop.Position.Y
	mag := math.Sqrt(dirX*dirX + dirY*dirY)

	if mag == 0 {
		return
	}

	// Chuẩn hóa vector
	dirX /= mag
	dirY /= mag

	// Cập nhật vị trí
	troop.Position.X += dirX * speed
	troop.Position.Y += dirY * speed
}

// moveAwayFromPosition - Di chuyển ra xa khỏi một vị trí
func (g *Game) moveAwayFromPosition(troop *model.TroopInstance, pos model.Position, speed float64) {
	dirX := troop.Position.X - pos.X
	dirY := troop.Position.Y - pos.Y
	mag := math.Sqrt(dirX*dirX + dirY*dirY)

	if mag == 0 {
		// Nếu trùng vị trí, di chuyển random
		dirX = 1.0
		dirY = 0.0
		mag = 1.0
	}

	// Chuẩn hóa vector và di chuyển ra xa
	dirX /= mag
	dirY /= mag

	newX := troop.Position.X + dirX*speed
	newY := troop.Position.Y + dirY*speed

	if g.isValidPosition(newX, newY) && !g.checkCollision(troop, newX, newY) {
		troop.Position.X = newX
		troop.Position.Y = newY
	}
}

// moveInSameDirection - Di chuyển cùng hướng với ally
func (g *Game) moveInSameDirection(healer *model.TroopInstance, ally *model.TroopInstance, speed float64) {
	if healer == nil || ally == nil {
		return
	}

	// Ước tính hướng di chuyển của ally dựa trên vị trí hiện tại
	isPlayer1 := healer.Owner == g.Player1.User.Username
	directionY := getDirectionY(isPlayer1)

	// Di chuyển song song với ally
	offsetX := 0.0 // Có thể thêm offset nhỏ để không đi trùng
	if healer.Position.X < ally.Position.X {
		offsetX = -0.5
	} else {
		offsetX = 0.5
	}

	newX := healer.Position.X + offsetX*speed*0.3
	newY := healer.Position.Y + directionY*speed

	if g.isValidPosition(newX, newY) && !g.checkCollision(healer, newX, newY) {
		healer.Position.X = newX
		healer.Position.Y = newY
	}
}

// moveHealerBackToSafety - Di chuyển healer về vùng an toàn qua cầu
func (g *Game) moveHealerBackToSafety(healer *model.TroopInstance, speed float64, isPlayer1 bool) {
	currentX := healer.Position.X
	currentY := healer.Position.Y
	directionY := -getDirectionY(isPlayer1)

	safeZoneY := getSafeZoneY(isPlayer1)

	if g.isInSafeZone(currentY, safeZoneY, isPlayer1) {
		g.waitForAlliesAtSafeZone(healer, speed*0.8, isPlayer1)
		return
	}

	if g.isInRiver(currentY) {
		g.handleHealerInRiver(healer, currentX, speed, directionY)
	} else {
		g.handleHealerOutsideRiver(healer, currentX, speed, directionY)
	}

	fmt.Printf("Healer %s retreating to safety at (%.1f, %.1f)\n",
		healer.Template.Name, healer.Position.X, healer.Position.Y)
}

func (g *Game) handleHealerInRiver(healer *model.TroopInstance, currentX, speed, directionY float64) {
	if !isBridgeColumn(currentX) {
		closestBridge := closestBridgeColumn(currentX)
		if currentX < closestBridge {
			healer.Position.X += min(speed, closestBridge-currentX)
		} else if currentX > closestBridge {
			healer.Position.X -= min(speed, currentX-closestBridge)
		}
	} else {
		healer.Position.Y += directionY * speed
	}
}

func (g *Game) handleHealerOutsideRiver(healer *model.TroopInstance, currentX, speed, directionY float64) {
	closestBridge := closestBridgeColumn(currentX)
	dx := closestBridge - currentX

	if utils.AbsFloat(dx) > 0.05 {
		if dx > 0 {
			healer.Position.X += min(speed*0.7, dx)
		} else {
			healer.Position.X += max(-speed*0.7, dx)
		}
		healer.Position.Y += directionY * speed * 0.5
	} else {
		healer.Position.X = closestBridge
		healer.Position.Y += directionY * speed
	}
}

// waitForAlliesAtSafeZone - Chờ đồng minh tại vùng an toàn
func (g *Game) waitForAlliesAtSafeZone(healer *model.TroopInstance, speed float64, isPlayer1 bool) {
	// Di chuyển về trung tâm map để dễ gặp đồng minh
	centerX := MAP_SIZE / 2

	if healer.Position.X < centerX-1 {
		healer.Position.X += speed
	} else if healer.Position.X > centerX+1 {
		healer.Position.X -= speed
	}

	// Duy trì vị trí Y trong vùng an toàn
	var idealY float64
	if isPlayer1 {
		idealY = 8.0 // Player 1 chờ ở Y = 7
	} else {
		idealY = 13.0 // Player 2 chờ ở Y = 14
	}

	if utils.AbsFloat(healer.Position.Y-idealY) > 0.5 {
		if healer.Position.Y < idealY {
			healer.Position.Y += speed * 0.5
		} else {
			healer.Position.Y -= speed * 0.5
		}
	}
}

// =============================================================================
// 3. HỆ THỐNG HỒI MÁU
// =============================================================================

// healAlly - Hồi máu cho đồng minh
func (g *Game) healAlly(healer *model.TroopInstance, target *model.TroopInstance) {
	if healer == nil || target == nil || healer.Template == nil || target.Template == nil {
		return
	}

	currentTime := time.Now()
	healCooldown := time.Duration(healer.Template.AttackSpeed * float64(time.Second))

	if currentTime.Sub(healer.LastAttackTime) < healCooldown {
		return
	}

	healAmount := healer.Template.DMG

	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	target.Template.HP += healAmount
	if target.Template.HP > target.Template.MaxHP {
		target.Template.HP = target.Template.MaxHP
	}

	healer.LastAttackTime = currentTime

	fmt.Printf("Healer %s healed ally %s for %.2f HP (%.1f/%.1f)\n",
		healer.Template.Name,
		target.Template.Name,
		healAmount,
		target.Template.HP,
		target.Template.MaxHP)
}





