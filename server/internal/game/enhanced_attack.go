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
func (g *Game) hasEnemyInRange(troop *model.TroopInstance) bool {
	if troop == nil || troop.Template == nil {
		return false
	}

	for _, entities := range g.BattleSystem.BattleMap { // Duyệt tất cả entity trên map
		for _, entity := range entities {
			otherTroop, ok := entity.(*model.TroopInstance)
			if !ok || otherTroop == nil || otherTroop.IsDead || otherTroop.Owner == troop.Owner {
				continue // Bỏ qua nếu không phải troop hoặc cùng phe hoặc đã chết
			}

			dist := calculateDistance(troop.Position, otherTroop.Position)
			if dist <= troop.Template.Range {
				return true // Tìm thấy ít nhất 1 con enemy trong range
			}
		}
	}
	return false // Không có con nào trong tầm
}

// getClosestEnemy - Lấy enemy troop gần nhất (không cần trong tầm đánh)
func (g *Game) getClosestEnemy(troop *model.TroopInstance) *model.TroopInstance {
	if troop == nil {
		return nil
	}

	var closestEnemy *model.TroopInstance // Biến lưu con enemy gần nhất
	minDist := math.MaxFloat64            // Khoảng cách nhỏ nhất ban đầu set là vô cực

	for _, entities := range g.BattleSystem.BattleMap { // Duyệt từng entity trên bản đồ
		for _, entity := range entities {
			otherTroop, ok := entity.(*model.TroopInstance)
			if !ok || otherTroop == nil || otherTroop.IsDead || otherTroop.Owner == troop.Owner {
				continue // Bỏ qua nếu không phải troop, đã chết hoặc cùng phe
			}

			dist := calculateDistance(troop.Position, otherTroop.Position) // Tính khoảng cách giữa 2 troop
			if dist < minDist {                                            // Nếu khoảng cách này nhỏ hơn khoảng cách trước đó
				closestEnemy = otherTroop // Cập nhật con enemy gần nhất
				minDist = dist
			}
		}
	}

	return closestEnemy // Trả về con enemy gần nhất
}

// findClosestEnemyInRange - Tìm enemy troop gần nhất trong phạm vi tấn công
func (g *Game) findClosestEnemyInRange(troop *model.TroopInstance) (bool, *model.TroopInstance, float64) {
	if troop == nil || troop.Template == nil {
		return false, nil, 0
	}

	enemyInRange := false
	var closestEnemy *model.TroopInstance
	minDist := math.MaxFloat64

	for _, entities := range g.BattleSystem.BattleMap {
		// Chỉ xét các troop instance
		for _, entity := range entities {
			e, ok := entity.(*model.TroopInstance)
			if !ok ||
				e.Owner == troop.Owner || // Bỏ qua troop cùng phe
				!e.IsAlive() { // Bỏ qua troop đã chết
				continue
			}

			// Tính khoảng cách đến troop địch
			distance := calculateDistance(troop.Position, e.Position)

			// Kiểm tra trong tầm và gần hơn target hiện tại
			if distance <= troop.Template.Range && distance < minDist {
				enemyInRange = true
				closestEnemy = e
				minDist = distance
			}
		}
	}

	return enemyInRange, closestEnemy, minDist
}

// CanAttackTower - Kiểm tra troop có thể tấn công tower không
func (g *Game) canAttackTower(troop *model.TroopInstance) (bool, *model.TowerInstance, float64) {
	if troop == nil || troop.Template == nil {
		return false, nil, 0
	}

	var closestTower *model.TowerInstance
	minDist := math.MaxFloat64

	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			// Chỉ xét các tower instance
			e, ok := entity.(*model.TowerInstance)
			if !ok ||
				e.Owner == troop.Owner || // Bỏ qua tower cùng phe
				!e.IsAlive() { // Bỏ qua tower đã bị phá
				continue
			}

			// Tính khoảng cách đến tower địch
			// towerPos := e.GetPosition()
			dist := calculateDistanceToTower(troop.Position, e.Area)

			// Kiểm tra trong tầm và gần hơn target hiện tại
			if dist <= troop.Template.Range && dist < minDist {
				closestTower = e
				minDist = dist
			}
		}
	}

	return closestTower != nil, closestTower, minDist
}

// =============================================================================
// 2. HÀNH VI DI CHUYỂN CỦA TROOP
// =============================================================================

// handleMovement - Xử lý di chuyển chính của troop
func (g *Game) handleMovement(troop *model.TroopInstance, moveSpeed, directionY float64, isPlayer1 bool) {
	if troop == nil {
		return
	}

	newX, newY := g.calculateNextPosition(troop, moveSpeed, directionY, isPlayer1)
	if !g.checkCollision(troop, newX, newY) && g.isValidPosition(newX, newY) {
		troop.Position.X = newX
		troop.Position.Y = newY
	} else {
		g.handleCollisionMovement(troop, newX, newY, moveSpeed)
	}
}

// handleCombatMovement - Xử lý di chuyển trong combat
func (g *Game) handleCombatMovement(troop, closestEnemy *model.TroopInstance, minDist, moveSpeed float64) {
	if troop == nil || closestEnemy == nil || troop.Template == nil {
		return
	}

	if minDist < troop.Template.Range*0.5 {
		// Tính hướng lùi lại từ enemy
		dx := troop.Position.X - closestEnemy.Position.X
		dy := troop.Position.Y - closestEnemy.Position.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > 0 {
			// Lùi lại một chút để tránh quá gần
			newX := troop.Position.X + (dx/dist)*moveSpeed*0.5
			newY := troop.Position.Y + (dy/dist)*moveSpeed*0.5

			if !g.checkCollision(troop, newX, newY) && g.isValidPosition(newX, newY) {
				troop.Position.X = newX
				troop.Position.Y = newY
			}
		}
	}
}

// calculateNextPosition - Tính toán vị trí tiếp theo dựa trên trạng thái hiện tại
func (g *Game) calculateNextPosition(troop *model.TroopInstance, moveSpeed, directionY float64, isPlayer1 bool) (float64, float64) {
	if troop == nil {
		return 0, 0
	}

	currentX := troop.Position.X
	currentY := troop.Position.Y

	enemy := g.getClosestEnemy(troop)
	if enemy != nil {
		dx := troop.Position.X - enemy.Position.X
		dy := troop.Position.Y - enemy.Position.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		rangeVal := troop.Template.Range
		if dist <= rangeVal {
			if dist < rangeVal*0.5 {
				// QUÁ GẦN: lùi nhẹ
				newX := troop.Position.X + (dx/dist)*moveSpeed*0.5
				newY := troop.Position.Y + (dy/dist)*moveSpeed*0.5

				if g.isValidPosition(newX, newY) && !g.checkCollision(troop, newX, newY) {
					return newX, newY
				}
			}
			// Trong tầm vừa đủ: đứng yên đánh
			return currentX, currentY
		}
	}

	// Lấy mục tiêu hiện tại
	var targetY float64
	if troop.Template.AggroPriority == "troop" {
		if enemy := g.getClosestEnemy(troop); enemy != nil {
			targetY = enemy.Position.Y
		} else {
			area := getTargetTowerArea(troop, g)
			targetY = (area.TopLeft.Y + area.BottomRight.Y) / 2
		}
	} else {
		area := getTargetTowerArea(troop, g)
		targetY = (area.TopLeft.Y + area.BottomRight.Y) / 2
	}

	// Nếu mục tiêu ở phía bên kia sông → đi tới cầu gần nhất
	if isTargetAcrossRiver(troop, targetY) {
		return g.moveTowardsBridge(currentX, currentY, moveSpeed, directionY)
	}

	// Kiểm tra trạng thái băng sông
	isNearRiver := (directionY > 0 && currentY < RIVER_TOP && currentY+moveSpeed >= RIVER_TOP) ||
		(directionY < 0 && currentY > RIVER_BOTTOM && currentY-moveSpeed <= RIVER_BOTTOM)
	isCrossingRiver := (currentY >= RIVER_TOP && currentY <= RIVER_BOTTOM)
	hasPassedRiver := (directionY > 0 && currentY > RIVER_BOTTOM) ||
		(directionY < 0 && currentY < RIVER_TOP)

	// Trước khi đến sông, di chuyển về phía cầu gần nhất
	if isNearRiver && !isCrossingRiver && !hasPassedRiver {
		if !isBridgeColumn(currentX) {
			// Chỉ move theo trục X nếu chưa đứng đúng cầu
			return g.strafeToBridge(currentX, currentY, moveSpeed)
		}
		// Nếu đã đứng đúng cầu thì đi chéo như thường
		return g.moveTowardsBridge(currentX, currentY, moveSpeed, directionY)
	}

	// Đang băng sông, giữ ở trên cầu
	if isCrossingRiver {
		return g.moveAcrossRiver(currentX, currentY, moveSpeed, directionY)
	}

	// Sau khi băng sông hoặc chưa đến sông, di chuyển về mục tiêu
	return g.moveTowardsTarget(troop, currentX, currentY, moveSpeed, isPlayer1)
}

func (g *Game) strafeToBridge(currentX, currentY, moveSpeed float64) (float64, float64) {
	closestBridge := closestBridgeColumn(currentX)
	dx := closestBridge - currentX

	var moveX float64
	if dx > 0 {
		moveX = min(moveSpeed, dx)
	} else {
		moveX = max(-moveSpeed, dx)
	}

	return currentX + moveX, currentY // Chỉ đổi X, không đổi Y
}

// moveTowardsBridge - Di chuyển về phía cầu gần nhất
func (g *Game) moveTowardsBridge(currentX, currentY, moveSpeed, directionY float64) (float64, float64) {
	closestBridge := closestBridgeColumn(currentX)

	// Vector hướng đến cầu
	dx := closestBridge - currentX
	dy := directionY * 1.0 // luôn muốn đi về phía sông theo Y

	// Chuẩn hóa vector
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag == 0 {
		return currentX, currentY
	}
	nx := dx / mag
	ny := dy / mag

	// Di chuyển theo hướng đó
	newX := currentX + nx*moveSpeed
	newY := currentY + ny*moveSpeed

	if newY >= RIVER_TOP && newY <= RIVER_BOTTOM {
		// Clamp lại X nếu đang ở trong sông
		newX = closestBridge*moveSpeed
	}

	return newX, newY
}

func isAcrossRiver(currentY, targetY float64) bool {
	// Một bên ở trên RIVER_TOP và một bên ở dưới RIVER_BOTTOM → khác phía
	return (currentY < RIVER_TOP && targetY > RIVER_BOTTOM) || (currentY > RIVER_BOTTOM && targetY < RIVER_TOP)
}

func isTargetAcrossRiver(troop *model.TroopInstance, targetY float64) bool {
	y := troop.Position.Y

	// Player 1 từ dưới lên, Player 2 từ trên xuống
	return (y < RIVER_TOP && targetY > RIVER_BOTTOM) || (y > RIVER_BOTTOM && targetY < RIVER_TOP)
}

// moveAcrossRiver - Di chuyển băng sông trên cầu
func (g *Game) moveAcrossRiver(currentX, currentY, moveSpeed, directionY float64) (float64, float64) {
	if isBridgeColumn(currentX) {
		// Đang ở trên cầu, tiến về phía trước
		return currentX, currentY + directionY*moveSpeed
	}

	// Không ở trên cầu, di chuyển về cầu gần nhất
	closestBridge := closestBridgeColumn(currentX)
	newX := currentX

	if currentX < closestBridge {
		newX = currentX + min(moveSpeed, closestBridge-currentX)
	} else if currentX > closestBridge {
		newX = currentX - min(moveSpeed, currentX-closestBridge)
	}

	return newX, currentY
}

// moveTowardsTarget - Di chuyển về phía mục tiêu (towers hoặc vị trí chiến lược)
func (g *Game) moveTowardsTarget(troop *model.TroopInstance, currentX, currentY, moveSpeed float64, isPlayer1 bool) (float64, float64) {
	if troop == nil || troop.Template == nil {
		return currentX, currentY
	}

	var targetArea model.Area

	// Xác định mục tiêu dựa trên aggro priority và tình huống hiện tại
	canAttackTower, _, _ := g.canAttackTower(troop)
	enemyInRange := g.hasEnemyInRange(troop)

	if troop.Template.AggroPriority == "tower" && canAttackTower {
		targetArea = getTargetTowerArea(troop, g)
	} else if troop.Template.AggroPriority == "troop" && enemyInRange {
		enemy := g.getClosestEnemy(troop)
		if enemy != nil {
			// Tạo khu vực nhỏ xung quanh enemy
			targetArea.TopLeft.X = enemy.Position.X - MIN_TROOP_DISTANCE
			targetArea.TopLeft.Y = enemy.Position.Y - MIN_TROOP_DISTANCE
			targetArea.BottomRight.X = enemy.Position.X + MIN_TROOP_DISTANCE
			targetArea.BottomRight.Y = enemy.Position.Y + MIN_TROOP_DISTANCE
		} else {
			targetArea = getTargetTowerArea(troop, g)
		}
	} else {
		targetArea = getTargetTowerArea(troop, g)
	}

	// Tính trung tâm mục tiêu
	targetCenterX := (targetArea.TopLeft.X + targetArea.BottomRight.X) / 2
	targetCenterY := (targetArea.TopLeft.Y + targetArea.BottomRight.Y) / 2

	// Nếu mục tiêu ở bên kia sông → đi về cầu gần nhất
	if isAcrossRiver(currentY, targetCenterY) {
		closestBridge := closestBridgeColumn(currentX)
		dx := closestBridge - currentX
		dy := targetCenterY - currentY
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 0 {
			dx /= dist
			dy /= dist
		}

		moveX := dx * moveSpeed
		moveY := dy * moveSpeed

		newX := currentX + moveX
		newY := currentY + moveY
		return newX, newY
	}

	// Nếu cùng phía sông → di chuyển thẳng
	dx := targetCenterX - currentX
	dy := targetCenterY - currentY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 0 {
		dx /= dist
		dy /= dist
	}

	// Áp dụng di chuyển với tránh khu vực tower
	moveX := dx * moveSpeed
	moveY := dy * moveSpeed

	newX := currentX + moveX
	newY := currentY + moveY

	// Tránh khu vực tower nếu không nhắm mục tiêu chúng trực tiếp
	if !canAttackTower {
		newX, newY = g.avoidTowerAreas(newX, newY, isPlayer1)
	}

	return newX, newY
}

// =============================================================================
// 3. HỆ THỐNG TẤN CÔNG
// =============================================================================

// attackTroop - Xử lý troop tấn công troop khác
func (g *Game) attackTroop(attacker *model.TroopInstance, target *model.TroopInstance) {
	if attacker == nil || target == nil || attacker.Template == nil || target.Template == nil {
		return
	}

	currentTime := time.Now()
	// Tính cooldown dựa trên attack speed (giây)
	attackCooldown := time.Duration(attacker.Template.AttackSpeed * float64(time.Second))

	// Kiểm tra cooldown tấn công
	if currentTime.Sub(attacker.LastAttackTime) < attackCooldown {
		return // Chưa đến lúc tấn công
	}

	// Kiểm tra lại sau khi lock
	if !attacker.IsAlive() || !target.IsAlive() {
		return
	}

	target.Mutex.Lock()
	defer target.Mutex.Unlock()

	damage := math.Max(attacker.Template.DMG, 1)
	target.Template.HP -= damage
	attacker.LastAttackTime = currentTime

	fmt.Printf("Troop %s attacks troop %s for %.1f damage. Target HP: %.1f\n",
		attacker.Template.Name, target.Template.Name, damage, target.Template.HP)

	// Kiểm tra target có chết không
	if target.Template.HP <= 0 {
		target.IsDead = true

		// Thưởng gold cho người tấn công (bằng EXP của target)
		if attacker.Owner == g.Player1.User.Username {
			g.Player1.Gold += target.Template.EXP
		} else {
			g.Player2.Gold += target.Template.EXP
		}

		// Thêm reward cho việc giết troop
		g.addKillReward(attacker.Owner, target)

		fmt.Printf("Troop %s killed!\n", target.Template.Name)
	}
}

// attackTower - Xử lý troop tấn công tower
func (g *Game) attackTower(troop *model.TroopInstance) {
	if troop == nil || troop.Template == nil {
		return
	}

	currentTime := time.Now()
	// Tính cooldown dựa trên attack speed (giây)
	attackCooldown := time.Duration(troop.Template.AttackSpeed * float64(time.Second))

	// Kiểm tra cooldown tấn công
	if currentTime.Sub(troop.LastAttackTime) < attackCooldown {
		return // Chưa đến lúc tấn công
	}

	_, closestTower, _ := g.canAttackTower(troop)
	if closestTower == nil || closestTower.Template == nil {
		return
	}

	// Lock tower để tránh race condition
	closestTower.Mutex.Lock()
	defer closestTower.Mutex.Unlock()

	// Kiểm tra lại sau khi lock
	if !closestTower.IsAlive() {
		return
	}

	damage := math.Max(troop.Template.DMG, 1)
	closestTower.Template.HP -= damage
	troop.LastAttackTime = currentTime

	fmt.Printf("Troop %s attacks tower %s for %.1f damage. Tower HP: %.1f\n",
		troop.Template.Name, closestTower.Template.Type, damage, closestTower.Template.HP)

	if closestTower.Template.HP <= 0 {
		fmt.Printf("Tower %s destroyed!\n", closestTower.Template.Type)
		g.addTowerDestroyReward(troop.Owner, closestTower)
		g.checkWinCondition()
	}
}