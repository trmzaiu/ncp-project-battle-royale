package game

import (
	"fmt"
	"log"
	"math"
	"royaka/internal/model"
	"royaka/internal/utils"
	"runtime/debug"
	"time"
)

// =============================================================================
// CONSTANTS
// =============================================================================

const (
	MAP_SIZE           = 21.0
	MIN_TROOP_DISTANCE = 0.8
	RIVER_TOP          = 9.0
	RIVER_BOTTOM       = 12.0
	BRIDGE_TOLERANCE   = 0.5
	COMBAT_SPEED_MULT  = 0.3
	TOWER_SPEED_MULT   = 0.5
)

var (
	BRIDGE_COLUMNS = []float64{4, 17}
)

// =============================================================================
// PHẦN 1: HÀM CHÍNH - ĐIỀU KHIỂN GAME LOOP
// =============================================================================

// UpdateBattleMap - Hàm chính update tất cả entities trong game mỗi frame
// Gọi hàm update tương ứng cho từng loại entity (troop hoặc tower)
func (g *Game) UpdateBattleMap() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] UpdateBattleMap crashed: %v", r)
			debug.PrintStack()
		}
	}()

	entities := g.BattleSystem.GetEntities()

	for _, entity := range entities {
		switch e := entity.(type) {
		case *model.TroopInstance:
			g.updateTroop(e)
		case *model.TowerInstance:
			g.updateTower(e)
		default:
			log.Printf("[WARN] Unknown entity type: %T", entity)
		}
	}
}

// =============================================================================
// PHẦN 2: LOGIC CỦA TROOP
// =============================================================================

// updateTroop - Logic chính của một troop trong mỗi frame
// Troop sẽ tìm target -> tấn công (nếu có) -> di chuyển (nếu không có target)
func (g *Game) updateTroop(troop *model.TroopInstance) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC][updateTroop] Troop ID: %s - Error: %v", troop.ID, r)
			debug.PrintStack()
		}
	}()

	if troop == nil || troop.Template == nil {
		return
	}

	// Lock troop để tránh race condition
	troop.Mutex.Lock()
	defer troop.Mutex.Unlock()

	if !troop.IsAlive() {
		return
	}

	// Xác định troop này thuộc Player 1 hay Player 2
	isPlayer1 := troop.Owner == g.Player1.User.Username
	directionY := getDirectionY(isPlayer1) // +1 hoặc -1 tùy hướng di chuyển

	// Nếu troop đã chạm tới cuối bản đồ phía bên kia thì dừng luôn
	if reachedMapEnd(isPlayer1, troop.Position.Y) {
		return
	}

	// Lấy tốc độ di chuyển cơ bản của troop
	speed := troop.Template.Speed

	// Tìm enemy gần nhất trong phạm vi tấn công
	enemyInRange, closestEnemy, minDist := g.findClosestEnemyInRange(troop)

	// Kiểm tra xem troop này có thể tấn công tower không
	canAttackTower, _, _ := g.canAttackTower(troop)

	// Quyết định ưu tiên tấn công troop hay tower dựa trên AggroPriority
	shouldAttackTroop, shouldAttackTower := decideAttackTargets(troop.Template.AggroPriority, enemyInRange, canAttackTower)

	// Tùy vào trạng thái combat, điều chỉnh tốc độ di chuyển cho hợp lý
	moveSpeed := adjustMoveSpeed(speed, shouldAttackTroop, shouldAttackTower)

	// Nếu nên tấn công troop và có enemy gần nhất
	if shouldAttackTroop && closestEnemy != nil {
		// Tấn công enemy
		g.attackTroop(troop, closestEnemy)
		// Di chuyển combat nếu cần (ví dụ: tiến lại gần 1 tí, hoặc dừng lại)
		g.handleCombatMovement(troop, closestEnemy, minDist, moveSpeed)
	}

	// Nếu nên tấn công tower thì xử lý luôn
	if shouldAttackTower {
		g.attackTower(troop)
	}

	// Nếu không đánh troop hoặc enemy còn xa, thì tiếp tục tiến về phía trước
	if !shouldAttackTroop || minDist >= troop.Template.Range*0.5 {
		g.handleMovement(troop, moveSpeed, directionY, isPlayer1)
	}

	// Đảm bảo vị trí không vượt quá giới hạn bản đồ (0 -> 21)
	troop.Position.X = utils.ClampFloat(troop.Position.X, 0, MAP_SIZE)
	troop.Position.Y = utils.ClampFloat(troop.Position.Y, 0, MAP_SIZE)
}

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

// =============================================================================
// PHẦN 3: HỆ THỐNG TÌM KIẾM TARGET
// =============================================================================

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
			distance := calculateDistance(troop.Position, e.GetPosition())

			// Kiểm tra trong tầm và gần hơn target hiện tại
			if distance <= troop.Template.Range && distance < minDist {
				closestTower = e
				minDist = distance
				return true, closestTower, minDist
			}
		}
	}

	return false, nil, minDist
}

// =============================================================================
// PHẦN 4: HỆ THỐNG TẤN CÔNG
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

// =============================================================================
// PHẦN 5: HỆ THỐNG DI CHUYỂN
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
			newX := troop.Position.X + (dx/dist)*moveSpeed*0.3
			newY := troop.Position.Y + (dy/dist)*moveSpeed*0.3

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

	// Ranh giới sông
	riverTop := 9.0
	riverBottom := 12.0

	// Kiểm tra trạng thái băng sông
	isNearRiver := (directionY > 0 && currentY < riverTop && currentY+moveSpeed >= riverTop) ||
		(directionY < 0 && currentY > riverBottom && currentY-moveSpeed <= riverBottom)
	isCrossingRiver := (currentY >= riverTop && currentY <= riverBottom)
	hasPassedRiver := (directionY > 0 && currentY > riverBottom) ||
		(directionY < 0 && currentY < riverTop)

	// Trước khi đến sông, di chuyển về phía cầu gần nhất
	if isNearRiver && !isCrossingRiver && !hasPassedRiver {
		return g.moveTowardsBridge(currentX, currentY, moveSpeed, directionY)
	}

	// Đang băng sông, giữ ở trên cầu
	if isCrossingRiver {
		return g.moveAcrossRiver(currentX, currentY, moveSpeed, directionY)
	}

	// Sau khi băng sông hoặc chưa đến sông, di chuyển về mục tiêu
	return g.moveTowardsTarget(troop, currentX, currentY, moveSpeed, isPlayer1)
}

// moveTowardsBridge - Di chuyển về phía cầu gần nhất
func (g *Game) moveTowardsBridge(currentX, currentY, moveSpeed, directionY float64) (float64, float64) {
	closestBridge := closestBridgeColumn(currentX)
	dx := closestBridge - currentX

	newX := currentX
	newY := currentY

	if utils.AbsFloat(dx) > 0.1 {
		// Di chuyển chéo về phía cầu
		if dx > 0 {
			newX = currentX + min(moveSpeed, dx)
		} else {
			newX = currentX + max(-moveSpeed, dx)
		}
		// Di chuyển Y chậm hơn khi đi tới cầu
		newY = currentY + directionY*moveSpeed*0.5
	} else {
		// Gần cầu, căn chỉnh và tiến về phía trước
		newX = closestBridge
		newY = currentY + directionY*moveSpeed
	}

	return newX, newY
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
			targetArea.TopLeft.X = enemy.Position.X - 0.5
			targetArea.TopLeft.Y = enemy.Position.Y - 0.5
			targetArea.BottomRight.X = enemy.Position.X + 0.5
			targetArea.BottomRight.Y = enemy.Position.Y + 0.5
		} else {
			targetArea = getTargetTowerArea(troop, g)
		}
	} else {
		targetArea = getTargetTowerArea(troop, g)
	}

	// Tính trung tâm mục tiêu
	targetCenterX := (targetArea.TopLeft.X + targetArea.BottomRight.X) / 2
	targetCenterY := (targetArea.TopLeft.Y + targetArea.BottomRight.Y) / 2

	// Tính vector di chuyển
	dx := targetCenterX - currentX
	dy := targetCenterY - currentY
	dist := math.Sqrt(dx*dx + dy*dy)

	// Chuẩn hóa hướng
	if dist > 0 {
		dx /= dist
		dy /= dist
	}

	// Áp dụng di chuyển với tránh khu vực tower
	moveX := dx * moveSpeed * 0.8
	moveY := dy * moveSpeed * 0.8

	newX := currentX + moveX
	newY := currentY + moveY

	// Tránh khu vực tower nếu không nhắm mục tiêu chúng trực tiếp
	if !canAttackTower {
		newX, newY = g.avoidTowerAreas(newX, newY, isPlayer1)
	}

	return newX, newY
}

// =============================================================================
// PHẦN 6: HỆ THỐNG VA CHẠM
// =============================================================================

// CheckCollision - Kiểm tra va chạm với các troop khác
func (g *Game) checkCollision(movingTroop *model.TroopInstance, newX, newY float64) bool {
	if movingTroop == nil {
		return true
	}

	// Kiểm tra va chạm với tất cả troop khác
	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			if otherTroop, ok := entity.(*model.TroopInstance); ok {
				// Bỏ qua chính nó và troop đã chết
				if otherTroop.ID != movingTroop.ID && otherTroop.IsAlive() {
					dist := calculateDistance(model.Position{X: newX, Y: newY}, otherTroop.Position)
					if dist < MIN_TROOP_DISTANCE {
						return true // Có va chạm
					}
				}
			}
		}
	}
	return false // Không có va chạm
}

// HandleCollisionMovement - Xử lý di chuyển khi có va chạm
// Thử các hướng di chuyển thay thế
func (g *Game) handleCollisionMovement(troop *model.TroopInstance, intendedX, intendedY, moveSpeed float64) {
	if troop == nil {
		return
	}

	// Tính hướng di chuyển dự định
	dx := intendedX - troop.Position.X
	dy := intendedY - troop.Position.Y

	// Thử các lựa chọn di chuyển thay thế
	alternatives := []struct {
		x, y     float64
		priority int
	}{
		// Thử di chuyển vòng quanh chướng ngại vật
		{troop.Position.X - moveSpeed*0.4, intendedY, 1},
		{troop.Position.X + moveSpeed*0.4, intendedY, 1},
		{troop.Position.X, troop.Position.Y + dy*0.5, 2},
		{intendedX*0.5 + troop.Position.X*0.5, intendedY, 2},
		// Các lựa chọn chéo
		{troop.Position.X - moveSpeed*0.3, troop.Position.Y + moveSpeed*0.3, 3},
		{troop.Position.X + moveSpeed*0.3, troop.Position.Y + moveSpeed*0.3, 3},
		// Bước nhỏ về phía trước
		{troop.Position.X + dx*0.1, troop.Position.Y + dy*0.1, 4},
	}

	// Sắp xếp theo ưu tiên và thử từng lựa chọn
	for priority := 1; priority <= 4; priority++ {
		for _, alt := range alternatives {
			if alt.priority == priority {
				if !g.checkCollision(troop, alt.x, alt.y) && g.isValidPosition(alt.x, alt.y) {
					troop.Position.X = alt.x
					troop.Position.Y = alt.y
					return
				}
			}
		}
	}

	// Nếu tất cả lựa chọn đều thất bại, thử di chuyển tối thiểu
	if !g.checkCollision(troop, troop.Position.X+dx*0.05, troop.Position.Y+dy*0.05) {
		troop.Position.X += dx * 0.05
		troop.Position.Y += dy * 0.05
	}
}

// isValidPosition - Kiểm tra vị trí có hợp lệ không với logic được cải thiện
func (g *Game) isValidPosition(x, y float64) bool {
	// Check map boundaries
	if x < 0 || x > MAP_SIZE || y < 0 || y > MAP_SIZE {
		return false
	}

	// Check if in river but not on bridge
	if y > RIVER_TOP && y < RIVER_BOTTOM && !isBridgeColumn(x) {
		return false
	}

	return true
}

// Avoid tower areas when moving
func (g *Game) avoidTowerAreas(x, y float64, isPlayer1 bool) (float64, float64) {
	var towerAreas []model.Area

	// Duyệt toàn bộ entity trong BattleMap
	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			tower, ok := entity.(*model.TowerInstance)
			if !ok || tower.Owner == g.getPlayerID(isPlayer1) || !tower.IsAlive() {
				continue // Không phải tower, hoặc tower cùng phe, hoặc đã chết
			}
			towerAreas = append(towerAreas, tower.Area)
		}
	}

	// Kiểm tra vị trí có nằm trong vùng tower nào không
	for _, area := range towerAreas {
		if x >= area.TopLeft.X && x <= area.BottomRight.X &&
			y >= area.TopLeft.Y && y <= area.BottomRight.Y {

			// Move to nearest edge of the tower area
			distToLeft := x - area.TopLeft.X
			distToRight := area.BottomRight.X - x
			distToTop := y - area.TopLeft.Y
			distToBottom := area.BottomRight.Y - y

			minDist := min(min(distToLeft, distToRight), min(distToTop, distToBottom))

			if minDist == distToLeft {
				x = area.TopLeft.X - 0.5
			} else if minDist == distToRight {
				x = area.BottomRight.X + 0.5
			} else if minDist == distToTop {
				y = area.TopLeft.Y - 0.5
			} else {
				y = area.BottomRight.Y + 0.5
			}
			break
		}
	}

	return x, y
}

// =============================================================================
// PHẦN 7: LOGIC CỦA TOWER
// =============================================================================

// updateTower - Logic chính của tower trong mỗi frame
// Tower sẽ tìm troop địch gần nhất trong tầm và tấn công
func (g *Game) updateTower(tower *model.TowerInstance) {
	if tower == nil || tower.Template == nil {
		return
	}

	tower.Mutex.Lock()
	defer tower.Mutex.Unlock()

	// Kiểm tra tower còn hoạt động không
	if !tower.IsAlive() || tower.Template.HP <= 0 {
		return
	}

	// Tìm troop địch gần nhất trong tầm tấn công
	target := g.findClosestEnemyTroop(tower)
	if target != nil {
		g.towerAttackTroop(tower, target)
	}
}

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
		fmt.Printf("Troop %s killed by tower!\n", target.Template.Name)
	}
}

// =============================================================================
// PHẦN 8: HỆ THỐNG THƯỞNG VÀ CHIẾN THẮNG
// =============================================================================

// AddKillReward - Thêm phần thưởng khi giết troop
func (g *Game) addKillReward(playerName string, killedTroop *model.TroopInstance) {
	reward := killedTroop.Template.EXP // Thưởng bằng một nửa EXP của troop bị giết

	if playerName == g.Player1.User.Username {
		g.Player1.Gold += reward
	} else {
		g.Player2.Gold += reward
	}
}

// AddTowerDestroyReward - Thêm phần thưởng khi phá tower
func (g *Game) addTowerDestroyReward(playerName string, killedTower *model.TowerInstance) {
	reward := killedTower.Template.EXP

	if playerName == g.Player1.User.Username {
		g.Player1.Gold += reward
	} else {
		g.Player2.Gold += reward
	}
}

// CheckWinCondition - Kiểm tra điều kiện thắng
func (g *Game) checkWinCondition() {
	// Chuyển gold cho users (lưu vào database)
	g.Player1.User.Gold += g.Player1.Gold
	g.Player2.User.Gold += g.Player2.Gold

	// Kiểm tra điều kiện thắng: phá được King Tower
	if g.Player2.Towers["king"].HP <= 0 {
		fmt.Printf("Player 1 (%s) wins by destroying the King Tower!\n", g.Player1.User.Username)
		return
	}

	if g.Player1.Towers["king"].HP <= 0 {
		fmt.Printf("Player 2 (%s) wins by destroying the King Tower!\n", g.Player2.User.Username)
		return
	}
}

// =============================================================================
// PHẦN 9: HÀM TIỆN ÍCH
// =============================================================================

// getDirectionY - Lấy hướng Y dựa trên player (1 hoặc -1)
func getDirectionY(isPlayer1 bool) float64 {
	if isPlayer1 {
		return 1.0
	}
	return -1.0
}

// reachedMapEnd - Kiểm tra troop đã đến cuối bản đồ chưa
func reachedMapEnd(isPlayer1 bool, y float64) bool {
	return (isPlayer1 && y >= 21.0) || (!isPlayer1 && y <= 0.0)
}

// decideAttackTargets - Quyết định mục tiêu tấn công dựa trên aggro priority
func decideAttackTargets(aggroPriority string, enemyInRange, canAttackTower bool) (bool, bool) {
	switch aggroPriority {
	case "troop":
		if enemyInRange {
			return true, false
		} else if canAttackTower {
			return false, true
		}
	case "tower":
		if canAttackTower {
			return false, true
		} else if enemyInRange {
			return true, false
		}
	default:
		return enemyInRange, canAttackTower && !enemyInRange
	}
	return false, false
}

// adjustMoveSpeed - Điều chỉnh tốc độ di chuyển dựa trên trạng thái combat
func adjustMoveSpeed(speed float64, shouldAttackTroop, shouldAttackTower bool) float64 {
	if shouldAttackTroop {
		return speed * 0.3
	} else if shouldAttackTower {
		return speed * 0.5
	}
	return speed
}

func getTargetTowerArea(troop *model.TroopInstance, g *Game) model.Area {
	isPlayer1 := troop.Owner == g.Player1.User.Username
	targetOwner := g.getPlayerID(!isPlayer1)

	var guard1, guard2, king *model.TowerInstance

	// Duyệt BattleMap để tìm các tower của đối thủ
	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			tower, ok := entity.(*model.TowerInstance)
			if !ok || tower.Owner != targetOwner || !tower.IsAlive() {
				continue
			}

			switch tower.Template.Type {
			case "guard1":
				guard1 = tower
			case "guard2":
				guard2 = tower
			case "king":
				king = tower
			}
		}
	}

	// Ưu tiên guard gần lane
	if troop.Position.X < 10 && guard1 != nil {
		return guard1.Area
	}
	if troop.Position.X >= 10 && guard2 != nil {
		return guard2.Area
	}

	// Nếu không có guard còn sống => king
	if king != nil {
		return king.Area
	}

	// Nếu tất cả đều null (có thể do lỗi) → fallback tránh panic
	return model.Area{
		TopLeft:     model.Position{X: 9, Y: 0},
		BottomRight: model.Position{X: 11, Y: 3},
	}
}

// calculateDistance - Tính khoảng cách Euclidean giữa 2 điểm
func calculateDistance(pos1, pos2 model.Position) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// isBridgeColumn - Kiểm tra có phải cột cầu không với tolerance được điều chỉnh
func isBridgeColumn(x float64) bool { // Vị trí 2 cầu
	bridgeTolerance := 0.5         // Tăng tolerance để dễ dàng đi qua cầu hơn

	for _, col := range BRIDGE_COLUMNS {
		if math.Abs(x-col) <= bridgeTolerance {
			return true
		}
	}
	return false
}

// Tìm cầu gần nhất với troop (cột X)
func closestBridgeColumn(x float64) float64 {
	closest := BRIDGE_COLUMNS[0]
	minDist := utils.AbsFloat(x - closest)

	for _, col := range BRIDGE_COLUMNS {
		dist := utils.AbsFloat(x - col)
		if dist < minDist {
			minDist = dist
			closest = col
		}
	}

	return closest
}

func (g *Game) getPlayerID(isPlayer1 bool) string {
	if isPlayer1 {
		return g.Player1.User.Username
	}
	return g.Player2.User.Username
}
