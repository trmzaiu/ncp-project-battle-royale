package game

import (
	"log"
	"royaka/internal/utils"
)

func (g *Game) HandleTurnTimeout() {
	log.Printf("[TURN] %s skipped due to timeout", g.Turn)

	nextPlayer := g.CurrentPlayer()
	g.SkipTurn(nextPlayer)

	log.Printf("[DEBUG][SKIP_TURN] Turn switched to: %s", g.Turn)

	payload := utils.Response{
		Type:    "skip_turn_response",
		Success: true,
		Message: "Turn skipped",
		Data: map[string]interface{}{
			"turn":    g.Turn,
			"player1": g.Player1,
			"player2": g.Player2,
		},
	}

	sendToClient(g.Player1.User.Username, payload)
	sendToClient(g.Player2.User.Username, payload)
}