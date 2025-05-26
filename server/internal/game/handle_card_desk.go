package game

import (
	"encoding/json"
	"log"
	"royaka/internal/model"
	"royaka/internal/utils"

	"github.com/gorilla/websocket"
)

func HandleGetDesk(conn *websocket.Conn, data json.RawMessage) {
	troops, err := model.LoadTroop()
	if err != nil {
		log.Println("loadTroop error:", err)
		conn.WriteJSON(utils.Response{
			Type:    "deck_response",
			Success: false,
			Message: "Failed to load troops",
		})
		return
	}

	conn.WriteJSON(utils.Response{
		Type:    "deck_response",
		Success: true,
		Message: "Troop data loaded",
		Data:    troops,
	})
}
