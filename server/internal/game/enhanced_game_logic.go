package game

import (
	"log"
	"math"
	"royaka/internal/model"
	"royaka/internal/utils"
	"runtime/debug"
)

// =============================================================================
// CONSTANTS
// =============================================================================

const (
	MAP_SIZE           = 21.0
	MIN_TROOP_DISTANCE = 0.3
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

	if troop.Template.Type == "healer" {
		g.updateHealerTroop(troop, isPlayer1)
		return
	}

	directionY := getDirectionY(isPlayer1) // +1 hoặc -1 tùy hướng di chuyển

	// Nếu troop đã chạm tới cuối bản đồ phía bên kia thì dừng luôn
	if reachedMapEnd(isPlayer1, troop.Position.Y) {
		return
	}

	// Lấy tốc độ di chuyển cơ bản của troop
	speed := troop.Template.Speed * 0.1

	// Tìm enemy gần nhất trong phạm vi tấn công
	enemyInRange, closestEnemy, minDist := g.getClosestEnemyInRange(troop)

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

func (g *Game) updateHealerTroop(troop *model.TroopInstance, isPlayer1 bool) {
	if troop == nil || troop.Template == nil {
		return
	}

	speed := troop.Template.Speed * 0.1

	// Kiểm tra xem healer có đang ở phe địch không
	if g.isHealerInEnemyTerritory(troop, isPlayer1) {
		// Nếu ở phe địch và không có đồng minh gần, quay về
		allyNearby := g.findAllyInRange(troop, troop.Template.Range*2) // Tìm trong phạm vi rộng hơn
		if allyNearby == nil {
			g.moveHealerBackToSafety(troop, speed, isPlayer1)
			return
		}
	}

	// Tìm đồng minh gần nhất cần hồi máu
	allyNeedHeal := g.findLowestHPAllyInRange(troop)

	if allyNeedHeal != nil {
		// Có ally cần heal
		dist := calculateDistance(troop.Position, allyNeedHeal.Position)
		if dist <= troop.Template.Range {
			// Trong tầm heal -> heal luôn
			g.healAlly(troop, allyNeedHeal)
		} else {
			// Ngoài tầm -> di chuyển lại gần để heal
			g.moveTowardPosition(troop, allyNeedHeal.Position, speed*0.8)
		}
	} else {
		// Không có ally cần heal -> tìm ally để follow
		allyToFollow := g.findAllyToFollow(troop)
		if allyToFollow != nil {
			// Có ally để theo -> follow với khoảng cách an toàn
			g.followAlly(troop, allyToFollow, speed)
		} else {
			// Không có ally nào -> kiểm tra vị trí và quyết định hành động
			g.handleHealerWithoutAllies(troop, speed, isPlayer1)
		}
	}

	// Clamp lại vị trí
	troop.Position.X = utils.ClampFloat(troop.Position.X, 0, MAP_SIZE)
	troop.Position.Y = utils.ClampFloat(troop.Position.Y, 0, MAP_SIZE)
}

// =============================================================================
// PHẦN 3: LOGIC CỦA TOWER
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

// =============================================================================
// PHẦN 9: HỆ THỐNG VA CHẠM
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
	towerAreas := g.getEnemyTowerAreas(isPlayer1)
	return adjustPositionOutsideTowerAreas(x, y, towerAreas)
}

// Helper to get enemy tower areas
func (g *Game) getEnemyTowerAreas(isPlayer1 bool) []model.Area {
	var towerAreas []model.Area
	for _, entities := range g.BattleSystem.BattleMap {
		for _, entity := range entities {
			tower, ok := entity.(*model.TowerInstance)
			if ok && tower.Owner != g.getPlayerID(isPlayer1) && tower.IsAlive() {
				towerAreas = append(towerAreas, tower.Area)
			}
		}
	}
	return towerAreas
}

// Helper to adjust position if inside any tower area
func adjustPositionOutsideTowerAreas(x, y float64, towerAreas []model.Area) (float64, float64) {
	for _, area := range towerAreas {
		if x >= area.TopLeft.X && x <= area.BottomRight.X &&
			y >= area.TopLeft.Y && y <= area.BottomRight.Y {

			distToLeft := x - area.TopLeft.X
			distToRight := area.BottomRight.X - x
			distToTop := y - area.TopLeft.Y
			distToBottom := area.BottomRight.Y - y

			minDist := min(min(distToLeft, distToRight), min(distToTop, distToBottom))

			switch minDist {
			case distToLeft:
				x = area.TopLeft.X - 0.5
			case distToRight:
				x = area.BottomRight.X + 0.5
			case distToTop:
				y = area.TopLeft.Y - 0.5
			default:
				y = area.BottomRight.Y + 0.5
			}
			break
		}
	}
	return x, y
}

// =============================================================================
// PHẦN 9: HỆ THỐNG THƯỞNG VÀ CHIẾN THẮNG
// =============================================================================

// CheckWinCondition - Kiểm tra điều kiện thắng
func (g *Game) checkWinCondition() {
	winner, result := g.CheckWinner()
	if result == "" {
		return
	}
	gameOverPayload := utils.Response{
		Type:    "game_over_response",
		Success: true,
		Message: result,
		Data: map[string]interface{}{
			"winner": winner,
		},
	}
	sendToClient(g.Player1.User.Username, gameOverPayload)
	sendToClient(g.Player2.User.Username, gameOverPayload)
}

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

// =============================================================================
// PHẦN 10: HÀM TIỆN ÍCH
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
		return speed * 0.8
	} else if shouldAttackTower {
		return speed * 0.6
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

func getSafeZoneY(isPlayer1 bool) float64 {
	if isPlayer1 {
		return 8.0
	}
	return 13.0
}

// calculateDistance - Tính khoảng cách Euclidean giữa 2 điểm
func calculateDistance(pos1, pos2 model.Position) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func calculateDistanceToTower(troopPos model.Position, towerArea model.Area) float64 {
	// Tính khoảng cách từ troop đến edge gần nhất của tower area

	// Kiểm tra xem troop có ở trong area không
	if troopPos.X >= towerArea.TopLeft.X && troopPos.X <= towerArea.BottomRight.X &&
		troopPos.Y >= towerArea.TopLeft.Y && troopPos.Y <= towerArea.BottomRight.Y {
		return 0 // Troop đang ở trong tower area
	}

	// Tìm điểm gần nhất trên edge của rectangle
	var closestX, closestY float64

	// Clamp X coordinate
	if troopPos.X < towerArea.TopLeft.X {
		closestX = towerArea.TopLeft.X
	} else if troopPos.X > towerArea.BottomRight.X {
		closestX = towerArea.BottomRight.X
	} else {
		closestX = troopPos.X
	}

	// Clamp Y coordinate
	if troopPos.Y < towerArea.TopLeft.Y {
		closestY = towerArea.TopLeft.Y
	} else if troopPos.Y > towerArea.BottomRight.Y {
		closestY = towerArea.BottomRight.Y
	} else {
		closestY = troopPos.Y
	}

	// Tính khoảng cách Euclidean
	dx := troopPos.X - closestX
	dy := troopPos.Y - closestY
	return math.Sqrt(dx*dx + dy*dy)
}

// isBridgeColumn - Kiểm tra có phải cột cầu không với tolerance được điều chỉnh
func isBridgeColumn(x float64) bool { // Vị trí 2 cầu
	for _, col := range BRIDGE_COLUMNS {
		if math.Abs(x-col) <= 0.5 {
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
