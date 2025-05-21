package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"royaka/internal/model"
	"royaka/internal/utils"
	"time"
)

type Game struct {
	Player1        *model.Player
	Player2        *model.Player
	Turn           string
	Started        bool
	Enhanced       bool
	StartTime      time.Time
	MaxTime        time.Duration
	TickRate       float64
	LastTick       time.Time
	BattleMap      []BattleEntity
	TickerStopChan chan struct{}
}

type BattleEntity interface {
	GetID() string
	GetOwner() string
	GetType() string
	GetPosition() model.Position
	IsAlive() bool
}

// ===================== Game Initialization =====================

func NewGame(p1, p2 *model.Player, mode string) *Game {
	if mode != "simple" && mode != "enhanced" {
		log.Fatal("Invalid game mode")
	}

	startingPlayer := p1.User.Username
	if rand.Intn(2) == 0 {
		startingPlayer = p2.User.Username
	}

	p1.LastManaRegen = time.Now()
	p2.LastManaRegen = time.Now()

	p1.TowerInstances = model.CreateTowerInstances(p1.Towers, p1.User.Username, true)
	p2.TowerInstances = model.CreateTowerInstances(p2.Towers, p2.User.Username, false)

	game := &Game{
		Player1:   p1,
		Player2:   p2,
		Turn:      startingPlayer,
		Started:   true,
		Enhanced:  (mode == "enhanced"),
		BattleMap: []BattleEntity{},
	}

	if game.Enhanced {
		game.StartTime = time.Now()
		game.MaxTime = 3 * time.Minute
		game.TickerStopChan = make(chan struct{})

		for _, t := range p1.TowerInstances {
			game.BattleMap = append(game.BattleMap, t)
		}
		for _, t := range p2.TowerInstances {
			game.BattleMap = append(game.BattleMap, t)
		}

		go game.startTicker()
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

	message := fmt.Sprintf("%s dealt %f damage to %s", troop.Name, damage, targetTower.Type)
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

	return int(damage), isCrit, message
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

	lowest.HP += float64(healAmount)
	if lowest.HP > lowest.MaxHP {
		lowest.HP = lowest.MaxHP
	}

	message := fmt.Sprintf("Queen healed %s tower for %f HP", lowest.Type, healAmount)
	if isCrit {
		message += " (Critical heal!)"
	}

	player.Turn++
	g.SwitchTurn()

	return int(healAmount), lowest, message
}

// ===================== Combat =====================

func (g *Game) AttackTower(player *model.Player, troop *model.Troop, tower *model.Tower) (float64, bool, bool) {
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
	g.StopGameLoop()
	if g.Player1.Towers["king"].HP <= 0.0 {
		g.Started = false
		if !g.Started {
			AwardEXP(g.Player2.User, g.Player1.User, false)
		}
		return g.Player2, g.Player2.User.Username + " wins!"
	}

	if g.Player2.Towers["king"].HP <= 0.0 {
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
	g.StopGameLoop()
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

// ===================== Game Tick =====================

func (g *Game) startTicker() {
	tickTicker := time.NewTicker(200 * time.Millisecond)    // Tick() mỗi 500ms
	updateTicker := time.NewTicker(1000 * time.Millisecond) // Update + Broadcast mỗi 1000ms
	defer tickTicker.Stop()
	defer updateTicker.Stop()

	for {
		select {
		case <-tickTicker.C:
			g.Tick()
		case <-updateTicker.C:
			g.UpdateBattleMap()
			g.BroadcastGameState()
		case <-g.TickerStopChan:
			log.Println("[INFO][GAME] Game loop stopped.")
			return
		}
	}
}

func (g *Game) StopGameLoop() {
	if g.Started {
		g.Started = false
		close(g.TickerStopChan)
	}
}

func (g *Game) Tick() {
	now := time.Now()

	// Apply mana regen for both players
	for _, player := range []*model.Player{g.Player1, g.Player2} {
		if player.Mana < 10 && now.Sub(player.LastManaRegen) >= 2*time.Second {
			player.Mana++
			player.LastManaRegen = now
			sendToClient(player.User.Username, utils.Response{
				Type:    "mana_update",
				Success: true,
				Message: fmt.Sprintf("Mana: %d", player.Mana),
				Data: map[string]interface{}{
					"player": player,
				},
			})
		}
	}
}

// ===================== Game Move =====================

func (g *Game) UpdateBattleMap() {
	for _, entity := range g.BattleMap {
		troop, ok := entity.(*model.TroopInstance)
		if !ok || troop == nil || troop.IsDead {
			continue
		}

		// Xác định hướng di chuyển dựa trên người chơi
		isPlayer1 := troop.Owner == g.Player1.User.Username
		directionY := 0.0
		if isPlayer1 {
			directionY = 1.0 // Player 1 đi từ dưới lên
		} else {
			directionY = -1.0 // Player 2 đi từ trên xuống
		}

		// Giới hạn biên bản đồ
		if (directionY > 0 && troop.Position.Y >= 20.0) || (directionY < 0 && troop.Position.Y <= 0.0) {
			continue
		}

		// Tốc độ di chuyển
		speed := troop.Template.Speed

		// Kiểm tra enemy trong tầm đánh
		enemyInRange := false
		// var closestEnemy *model.TroopInstance
		minDist := math.MaxFloat64
		
		for _, otherEntity := range g.BattleMap {
			otherTroop, isOtherTroop := otherEntity.(*model.TroopInstance)
			if !isOtherTroop || otherTroop == nil || otherTroop.IsDead || otherTroop.Owner == troop.Owner {
				continue
			}
			
			dist := calculateDistance(troop.Position, otherTroop.Position)
			if dist < troop.Template.Range && dist < minDist {
				enemyInRange = true
				// closestEnemy = otherTroop
				minDist = dist
			}
		}
		
		// Nếu có địch trong tầm đánh, dừng lại và không di chuyển
		if enemyInRange {
			continue
		}

		// Xác định vị trí sông
		riverTop := 9.0
		riverBottom := 11.0
		isNearRiver := (directionY > 0 && troop.Position.Y < riverTop && troop.Position.Y+speed >= riverTop) ||
			(directionY < 0 && troop.Position.Y > riverBottom && troop.Position.Y-speed <= riverBottom)
		isCrossingRiver := (troop.Position.Y >= riverTop && troop.Position.Y <= riverBottom)
		hasPassedRiver := (directionY > 0 && troop.Position.Y > riverBottom) || 
			(directionY < 0 && troop.Position.Y < riverTop)

		// Trước khi tới sông, di chuyển hướng tới cầu gần nhất
		if isNearRiver && !isCrossingRiver && !hasPassedRiver {
			// Tìm cầu gần nhất
			closestBridge := closestBridgeColumn(troop.Position.X)
			
			// Tính toán vector di chuyển tới cầu
			dx := closestBridge - troop.Position.X
			
			// Di chuyển xéo tới cầu
			moveX := 0.0
			if absFloat(dx) > 0.1 {
				// Chuẩn hóa dx
				if dx > 0 {
					moveX = min(speed * 0.8, dx)
				} else {
					moveX = max(-speed * 0.8, dx)
				}
				
				// Di chuyển chậm hơn theo chiều Y khi đang đi tới cầu
				moveY := directionY * speed * 0.5
				
				// Cập nhật vị trí
				troop.Position.X += moveX
				troop.Position.Y += moveY
			} else {
				// Đã gần cầu, di chuyển thẳng tới cầu
				troop.Position.X = closestBridge
				troop.Position.Y += directionY * speed
			}
		} else if isCrossingRiver {
			// Đang qua sông, chỉ di chuyển theo hướng Y nếu đang ở cầu
			if isBridgeColumn(troop.Position.X) {
				troop.Position.Y += directionY * speed
			} else {
				// Không ở cầu, tìm cầu gần nhất
				closestBridge := closestBridgeColumn(troop.Position.X)
				
				if troop.Position.X < closestBridge {
					troop.Position.X += min(speed, closestBridge-troop.Position.X)
				} else {
					troop.Position.X -= min(speed, troop.Position.X-closestBridge)
				}
			}
		} else {
			// Đã qua sông hoặc chưa gần sông, di chuyển tới tháp địch
			targetArea := getTargetTowerArea(troop, g)
			
			// Tính trung tâm mục tiêu
			targetCenterX := (targetArea.TopLeft.X + targetArea.BottomRight.X) / 2
			targetCenterY := (targetArea.TopLeft.Y + targetArea.BottomRight.Y) / 2
			
			// Tính vector hướng tới mục tiêu
			dx := targetCenterX - troop.Position.X
			dy := targetCenterY - troop.Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			
			// Normalize hướng
			if dist > 0 {
				dx /= dist
				dy /= dist
			}
			
			// Di chuyển theo hướng tới mục tiêu
			moveX := dx * speed * 0.8
			moveY := dy * speed * 0.8
			
			troop.Position.X += moveX
			troop.Position.Y += moveY
		}
		
		// Clamp vị trí trong giới hạn bản đồ
		troop.Position.X = clampFloat(troop.Position.X, 0, 20)
		troop.Position.Y = clampFloat(troop.Position.Y, 0, 20)
	}
}

// Tính khoảng cách giữa hai điểm
func calculateDistance(pos1, pos2 model.Position) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Kiểm tra troop ở cột cầu
func isBridgeColumn(x float64) bool {
	bridgeCols := []float64{4, 5, 15, 16}
	for _, col := range bridgeCols {
		if absFloat(x-col) < 0.5 { // Tăng dung sai để phù hợp hơn
			return true
		}
	}
	return false
}

// Tìm cầu gần nhất với troop (cột X)
func closestBridgeColumn(x float64) float64 {
	bridgeCols := []float64{4, 5, 15, 16}
	closest := bridgeCols[0]
	minDist := absFloat(x - closest)

	for _, col := range bridgeCols {
		dist := absFloat(x - col)
		if dist < minDist {
			minDist = dist
			closest = col
		}
	}

	return closest
}

func getTargetTowerArea(troop *model.TroopInstance, g *Game) model.Area {
	isPlayer1 := troop.Owner == g.Player1.User.Username

	// Ưu tiên guard gần lane
	if troop.Position.X < 10 {
		if (!isPlayer1 && g.Player1.Towers["guard1"].HP > 0.0) || (isPlayer1 && g.Player2.Towers["guard1"].HP > 0.0) {
			return model.GetTowerArea("guard1", !isPlayer1)
		}
	} else {
		if (!isPlayer1 && g.Player1.Towers["guard2"].HP > 0.0) || (isPlayer1 && g.Player2.Towers["guard2"].HP > 0.0) {
			return model.GetTowerArea("guard2", !isPlayer1)
		}
	}

	// Nếu cả hai guard bị phá => King
	return model.GetTowerArea("king", !isPlayer1)
}

func clampFloat(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// Hàm tính trị tuyệt đối float
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Hàm lấy giá trị nhỏ hơn
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Hàm lấy giá trị lớn hơn
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ===================== Game Utility =====================

func (g *Game) BroadcastGameState() {
	for _, player := range []*model.Player{g.Player1, g.Player2} {
		sendToClient(player.User.Username, utils.Response{
			Type:    "game_state",
			Success: true,
			Message: "Game updated",
			Data: map[string]interface{}{
				"battleMap": g.BattleMap,
				"player":    player,
				"opponent":  g.Opponent(player),
			},
		})
	}
}

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
