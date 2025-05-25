package game

// import (
// 	"fmt"
// 	"math"
// 	"royaka/internal/model"
// 	"royaka/internal/utils"
// 	"sync"
// 	"time"
// )

// // ========== MAIN FUNCTION ==========

// // Gọi update cho tất cả troops và towers trong game
// func (g *Game) UpdateBattleMa() {
// 	var wg sync.WaitGroup

// 	// Troops
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		for _, entity := range g.BattleMap {
// 			troop, ok := entity.(*model.TroopInstance)
// 			if !ok || troop == nil || troop.IsDead {
// 				continue
// 			}
// 			g.updateTroopState(troop)
// 		}
// 	}()

// 	// Towers
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		for _, entity := range g.BattleMap {
// 			tower, ok := entity.(*model.TowerInstance)
// 			if !ok || tower == nil || tower.IsDestroyed {
// 				continue
// 			}
// 			g.updateTowerState(tower)
// 		}
// 	}()

// 	wg.Wait()
// }

// // ========== TROOP LOGIC ==========

// // updateTroopState - Xử lý logic chính cho một troop trong battle map
// func (g *Game) updateTroopState(troop *model.TroopInstance) {
// 	// Xác định troop này thuộc Player 1 hay Player 2
// 	isPlayer1 := troop.Owner == g.Player1.User.Username
// 	directionY := getDirectionY(isPlayer1) // +1 hoặc -1 tùy hướng di chuyển

// 	// Nếu troop đã chạm tới cuối bản đồ phía bên kia thì dừng luôn
// 	if reachedMapEnd(isPlayer1, troop.Position.Y) {
// 		return
// 	}

// 	// Lấy tốc độ di chuyển cơ bản của troop
// 	speed := troop.Template.Speed

// 	// Tìm enemy gần nhất trong phạm vi tấn công
// 	enemyInRange, closestEnemy, minDist := g.findClosestEnemyInRange(troop)

// 	// Kiểm tra xem troop này có thể tấn công tower không
// 	canAttackTower := g.CanAttackTower(troop)

// 	// Quyết định ưu tiên tấn công troop hay tower dựa trên AggroPriority
// 	shouldAttackTroop, shouldAttackTower := decideAttackTargets(troop.Template.AggroPriority, enemyInRange, canAttackTower)

// 	// Tùy vào trạng thái combat, điều chỉnh tốc độ di chuyển cho hợp lý
// 	moveSpeed := adjustMoveSpeed(speed, shouldAttackTroop, shouldAttackTower)

// 	// Nếu nên tấn công troop và có enemy gần nhất
// 	if shouldAttackTroop && closestEnemy != nil {
// 		// Tấn công enemy
// 		g.attackTarget(troop, closestEnemy)
// 		// Di chuyển combat nếu cần (ví dụ: tiến lại gần 1 tí, hoặc dừng lại)
// 		g.handleCombatMovement(troop, closestEnemy, minDist, moveSpeed)
// 	}

// 	// Nếu nên tấn công tower thì xử lý luôn
// 	if shouldAttackTower {
// 		g.attackTower(troop)
// 	}

// 	// Nếu không đánh troop hoặc enemy còn xa, thì tiếp tục tiến về phía trước
// 	if !shouldAttackTroop || minDist >= troop.Template.Range*0.5 {
// 		g.handleMovement(troop, moveSpeed, directionY, isPlayer1)
// 	}

// 	// Đảm bảo vị trí không vượt quá giới hạn bản đồ (0 -> 21)
// 	troop.Position.X = utils.ClampFloat(troop.Position.X, 0, 21)
// 	troop.Position.Y = utils.ClampFloat(troop.Position.Y, 0, 21)
// }




// // ========== TOWER LOGIC ==========

// // updateTowerState - Cập nhật trạng thái của một tower
// func (g *Game) updateTowerState(tower *model.TowerInstance) {
// 	// Xác định tower này thuộc player nào
// 	isPlayer1Tower := tower.Owner == g.Player1.User.Username

// 	// Tìm enemy troop gần nhất trong tầm bắn
// 	closestEnemy, _ := g.findClosestEnemyTroopInRange(tower, isPlayer1Tower)

// 	// Nếu có enemy trong tầm, tấn công nó
// 	if closestEnemy != nil {
// 		g.towerAttackTarget(tower, closestEnemy)
// 	}
// }

// // towerAttackTarget - Xử lý tower tấn công một troop
// func (g *Game) towerAttackTarget(tower *model.TowerInstance, target *model.TroopInstance) {
// 	// Kiểm tra cooldown tấn công
// 	currentTime := time.Now()
// 	if currentTime.Sub(tower.LastAttackTime) < time.Duration(tower.Template.AttackSpeed*float64(time.Second)) {
// 		return // Chưa sẵn sàng tấn công
// 	}

// 	// Tính damage
// 	damage := tower.Template.ATK
// 	if damage < 1 {
// 		damage = 1 // Damage tối thiểu
// 	}

// 	// Gây damage cho target
// 	target.Template.HP -= damage
// 	tower.LastAttackTime = currentTime

// 	// Log việc tấn công
// 	fmt.Printf("Tower %s attacks troop %s for %.1f damage. Target HP: %.1f\n",
// 		tower.ID, target.ID, damage, target.Template.HP)

// 	// Kiểm tra nếu target bị giết
// 	if target.Template.HP <= 0 {
// 		target.IsDead = true
// 		g.RemoveDeadTroop(target)

// 		// Thêm phần thưởng cho chủ sở hữu tower
// 		g.AddKillReward(tower.Owner, target)

// 		fmt.Printf("Tower %s killed troop %s\n", tower.ID, target.ID)
// 	}
// }

// // ========== ATTACK LOGIC ==========

// // attackTarget - Xử lý troop tấn công troop khác
// func (g *Game) attackTarget(attacker, target *model.TroopInstance) {
// 	// Kiểm tra cooldown tấn công
// 	currentTime := time.Now()
// 	if currentTime.Sub(attacker.LastAttackTime) < time.Duration(attacker.Template.AttackSpeed*float64(time.Second)) {
// 		return // Chưa đến lúc tấn công
// 	}

// 	// Thực hiện tấn công
// 	damage := attacker.Template.DMG
// 	if damage < 1 {
// 		damage = 1 // Damage tối thiểu
// 	}

// 	target.Template.HP -= damage
// 	attacker.LastAttackTime = currentTime

// 	// Log tấn công
// 	fmt.Printf("Troop %s attacks %s for %.1f damage. Target HP: %.1f\n",
// 		attacker.ID, target.ID, damage, target.Template.HP)

// 	// Kiểm tra nếu mục tiêu chết
// 	if target.Template.HP <= 0 {
// 		target.IsDead = true
// 		g.RemoveDeadTroop(target)

// 		// Thêm EXP cho troop tấn công
// 		attacker.Template.EXP += target.Template.EXP / 2

// 		// Thêm phần thưởng cho người chơi
// 		g.AddKillReward(attacker.Owner, target)
// 	}
// }

// // CanAttackTower - Kiểm tra troop có thể tấn công tower không
// func (g *Game) CanAttackTower(troop *model.TroopInstance) bool {
// 	isPlayer1 := troop.Owner == g.Player1.User.Username
// 	var targetTowers map[string]*model.Tower

// 	if isPlayer1 {
// 		targetTowers = g.Player2.Towers
// 	} else {
// 		targetTowers = g.Player1.Towers
// 	}

// 	// Kiểm tra từng tower
// 	for towerID, tower := range targetTowers {
// 		if tower.HP <= 0 {
// 			continue
// 		}

// 		towerArea := model.GetTowerArea(towerID, !isPlayer1)
// 		towerCenterX := (towerArea.TopLeft.X + towerArea.BottomRight.X) / 2
// 		towerCenterY := (towerArea.TopLeft.Y + towerArea.BottomRight.Y) / 2

// 		towerPos := model.Position{X: towerCenterX, Y: towerCenterY}
// 		dist := calculateDistance(troop.Position, towerPos)

// 		if dist <= troop.Template.Range {
// 			return true
// 		}
// 	}

// 	return false
// }

// // attackTower - Xử lý troop tấn công tower
// func (g *Game) attackTowers(troop *model.TroopInstance) {
// 	// Kiểm tra cooldown tấn công
// 	currentTime := time.Now()
// 	if currentTime.Sub(troop.LastAttackTime) < time.Duration(troop.Template.AttackSpeed)*time.Millisecond {
// 		return
// 	}

// 	isPlayer1 := troop.Owner == g.Player1.User.Username
// 	var targetTowers map[string]*model.Tower

// 	if isPlayer1 {
// 		targetTowers = g.Player2.Towers
// 	} else {
// 		targetTowers = g.Player1.Towers
// 	}

// 	// Tìm tower gần nhất trong tầm đánh
// 	var closestTower *model.Tower
// 	var closestTowerID string
// 	minDist := math.MaxFloat64

// 	for towerID, tower := range targetTowers {
// 		if tower.HP <= 0 {
// 			continue
// 		}

// 		towerArea := model.GetTowerArea(towerID, !isPlayer1)
// 		towerCenterX := (towerArea.TopLeft.X + towerArea.BottomRight.X) / 2
// 		towerCenterY := (towerArea.TopLeft.Y + towerArea.BottomRight.Y) / 2

// 		towerPos := model.Position{X: towerCenterX, Y: towerCenterY}
// 		dist := calculateDistance(troop.Position, towerPos)

// 		if dist <= troop.Template.Range && dist < minDist {
// 			closestTower = tower
// 			closestTowerID = towerID
// 			minDist = dist
// 		}
// 	}

// 	// Tấn công tower gần nhất
// 	if closestTower != nil {
// 		damage := float64(troop.Template.DMG)
// 		closestTower.HP -= damage
// 		troop.LastAttackTime = currentTime

// 		fmt.Printf("Troop %s attacks tower %s for %.1f damage. Tower HP: %.1f\n",
// 			troop.ID, closestTowerID, damage, closestTower.HP)

// 		// Kiểm tra nếu tower bị phá
// 		if closestTower.HP <= 0 {
// 			fmt.Printf("Tower %s destroyed!\n", closestTowerID)

// 			// Thêm phần thưởng cho người chơi
// 			g.AddTowerDestroyReward(troop.Owner, closestTowerID)

// 			// Kiểm tra điều kiện thắng
// 			g.CheckWinCondition()
// 		}
// 	}
// }

// // ========== MOVEMENT FUNCTIONS ==========






// // ========== COLLISION LOGIC ==========

// // HandleCollisionMovement - Xử lý va chạm cải tiến với pathfinding tốt hơn
// func (g *Game) HandleCollisionMovement(troop *model.TroopInstance, intendedX, intendedY, moveSpeed float64) {
// 	// Tính hướng di chuyển dự định
// 	dx := intendedX - troop.Position.X
// 	dy := intendedY - troop.Position.Y

// 	// Thử các lựa chọn di chuyển thay thế
// 	alternatives := []struct {
// 		x, y     float64
// 		priority int
// 	}{
// 		// Thử di chuyển vòng quanh chướng ngại vật
// 		{troop.Position.X - moveSpeed*0.4, intendedY, 1},
// 		{troop.Position.X + moveSpeed*0.4, intendedY, 1},
// 		{troop.Position.X, troop.Position.Y + dy*0.5, 2},
// 		{intendedX*0.5 + troop.Position.X*0.5, intendedY, 2},
// 		// Các lựa chọn chéo
// 		{troop.Position.X - moveSpeed*0.3, troop.Position.Y + moveSpeed*0.3, 3},
// 		{troop.Position.X + moveSpeed*0.3, troop.Position.Y + moveSpeed*0.3, 3},
// 		// Bước nhỏ về phía trước
// 		{troop.Position.X + dx*0.1, troop.Position.Y + dy*0.1, 4},
// 	}

// 	// Sắp xếp theo ưu tiên và thử từng lựa chọn
// 	for priority := 1; priority <= 4; priority++ {
// 		for _, alt := range alternatives {
// 			if alt.priority == priority {
// 				if !g.CheckCollision(troop, alt.x, alt.y) && g.isValidPosition(alt.x, alt.y) {
// 					troop.Position.X = alt.x
// 					troop.Position.Y = alt.y
// 					return
// 				}
// 			}
// 		}
// 	}

// 	// Nếu tất cả lựa chọn đều thất bại, thử di chuyển tối thiểu
// 	if !g.CheckCollision(troop, troop.Position.X+dx*0.05, troop.Position.Y+dy*0.05) {
// 		troop.Position.X += dx * 0.05
// 		troop.Position.Y += dy * 0.05
// 	}
// }

// // ========== HELPER FUNCTIONS  ===========


// // Xóa troop đã chết khỏi battle map
// func (g *Game) RemoveDeadTroop(troop *model.TroopInstance) {
// 	for i, entity := range g.BattleMap {
// 		if t, ok := entity.(*model.TroopInstance); ok && t.ID == troop.ID {
// 			g.BattleMap = append(g.BattleMap[:i], g.BattleMap[i+1:]...)
// 			break
// 		}
// 	}
// }

// // Thêm phần thưởng khi giết troop
// func (g *Game) AddKillReward(playerName string, killedTroop *model.TroopInstance) {
// 	reward := killedTroop.Template.EXP // Thưởng bằng 1/2 cost của troop bị giết

// 	if playerName == g.Player1.User.Username {
// 		g.Player1.Gold += reward
// 	} else {
// 		g.Player2.Gold += reward
// 	}
// }

// // Thêm phần thưởng khi phá tower
// func (g *Game) AddTowerDestroyReward(playerName string, towerID string) {
// 	var reward int
// 	switch towerID {
// 	case "guard1", "guard2":
// 		reward = 100
// 	case "king":
// 		reward = 200
// 	default:
// 		reward = 50
// 	}

// 	if playerName == g.Player1.User.Username {
// 		g.Player1.Gold += reward
// 	} else {
// 		g.Player2.Gold += reward
// 	}
// }

// // Kiểm tra điều kiện thắng
// func (g *Game) Chen() {
// 	g.Player1.User.Gold += g.Player1.Gold
// 	g.Player2.User.Gold += g.Player2.Gold
// 	// Player 1 thắng nếu phá King tower của Player 2
// 	if g.Player2.Towers["king"].HP <= 0 {
// 		// g.GameState = "player1_win"
// 		fmt.Printf("Player 1 (%s) wins!\n", g.Player1.User.Username)
// 		return
// 	}

// 	// Player 2 thắng nếu phá King tower của Player 1
// 	if g.Player1.Towers["king"].HP <= 0 {
// 		// g.GameState = "player2_win"
// 		fmt.Printf("Player 2 (%s) wins!\n", g.Player2.User.Username)
// 		return
// 	}
// }

