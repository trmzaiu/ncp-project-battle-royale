package game

import (
	"fmt"
	"log"
	"math/rand"
	"royaka/internal/model"
	"royaka/internal/utils"
	"sync"
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
	LastTick       time.Time
	TickerStopChan chan struct{}
	BattleSystem   *BattleSystem
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

	// Combine all tower instances
	initialEntities := make(map[string][]BattleEntity)
	for _, ti := range p1.TowerInstances {
		posKey := ti.GetPosition().String()
		initialEntities[posKey] = append(initialEntities[posKey], ti)
	}
	for _, ti := range p2.TowerInstances {
		posKey := ti.GetPosition().String()
		initialEntities[posKey] = append(initialEntities[posKey], ti)
	}

	battleSystem := &BattleSystem{
		BattleMap:      initialEntities,
		MapMutex:       sync.RWMutex{},
		TickerStopChan: make(chan struct{}),
		TickRate:       100 * time.Millisecond, // hoặc tùy bạn
	}

	game := &Game{
		Player1:        p1,
		Player2:        p2,
		Turn:           startingPlayer,
		Started:        true,
		Enhanced:       (mode == "enhanced"),
		BattleSystem:   battleSystem,
		TickerStopChan: battleSystem.TickerStopChan,
	}

	if game.Enhanced {
		game.StartTime = time.Now()
		game.MaxTime = 3 * time.Minute
		time.AfterFunc(3*time.Second, func() {
			game.StartTime = time.Now()
			go game.startTicker()
		})
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

// ===================== Game Outcome =====================

func (g *Game) CheckWinner() (*model.Player, string) {

	if g.Enhanced {
		g.Player1.User.Gold += g.Player1.Gold
		g.Player2.User.Gold += g.Player2.Gold
	}

	if g.Player1.Towers["king"].HP <= 0.0 {
		AwardEXP(g.Player2.User, g.Player1.User, false)
		g.StopGameLoop()
		return g.Player2, g.Player2.User.Username + " wins!"
	}

	if g.Player2.Towers["king"].HP <= 0.0 {
		AwardEXP(g.Player1.User, g.Player2.User, false)
		g.StopGameLoop()
		return g.Player1, g.Player1.User.Username + " wins!"
	}

	if g.Enhanced && time.Since(g.StartTime) > g.MaxTime {
		p1Score := g.Player1.DestroyedCount()
		p2Score := g.Player2.DestroyedCount()

		if p1Score > p2Score {
			AwardEXP(g.Player1.User, g.Player2.User, false)
			g.StopGameLoop()
			return g.Player1, g.Player1.User.Username + " wins by score!"
		}
		if p2Score > p1Score {
			AwardEXP(g.Player2.User, g.Player1.User, false)
			g.StopGameLoop()
			return g.Player2, g.Player2.User.Username + " wins by score!"
		}

		AwardEXP(g.Player1.User, g.Player2.User, true)
		g.StopGameLoop()
		return nil, "It's a draw!"
	}

	return nil, ""
}

func (g *Game) SetWinner(winner *model.Player) {
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
	manaTicker := time.NewTicker(200 * time.Millisecond) // Cập nhật mana mỗi 200ms
	tickTicker := time.NewTicker(g.BattleSystem.TickRate)
	cleanupTicker := time.NewTicker(5 * time.Second) // Dọn dẹp mỗi 5 giây

	defer func() {
		tickTicker.Stop()
		cleanupTicker.Stop()
	}()

	for {
		select {
		case <-tickTicker.C:
			g.UpdateBattleMap()
			g.BroadcastGameState()
		case <-manaTicker.C:
			g.UpdateMana()
		case <-cleanupTicker.C:
			g.BattleSystem.CleanupDeadEntities()

		case <-g.BattleSystem.TickerStopChan:
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

func (g *Game) UpdateMana() {
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

// ===================== Game Utility =====================

func (g *Game) BroadcastGameState() {
	elapsed := time.Since(g.StartTime)

	timeLeft := g.MaxTime - elapsed
	if timeLeft < 0 {
		timeLeft = 0
	}

	for _, player := range []*model.Player{g.Player1, g.Player2} {
		sendToClient(player.User.Username, utils.Response{
			Type:    "game_state",
			Success: true,
			Message: "Game updated",
			Data: map[string]interface{}{
				"battleMap": g.BattleSystem.GetEntityList(),
				"timeLeft":  timeLeft.Milliseconds(),
			},
		})
	}

	if timeLeft == 0 {
		g.CheckWinner()
		g.StopGameLoop()
	}
}

func (g *Game) Opponent(p *model.Player) *model.Player {
	if g.Player1.User.Username == p.User.Username {
		return g.Player2
	}
	return g.Player1
}
