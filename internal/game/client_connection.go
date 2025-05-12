package game

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientConnection struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
}

func (c *ClientConnection) SafeWrite(data interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	if c.Conn == nil {
		log.Println("[WS] No connection to write to")
		return nil
	}

	err := c.Conn.WriteJSON(data)
	if err != nil {
		log.Printf("[WS] Failed to send: %v", err)
	} else {
		log.Printf("[WS] Sent: %T", data)
	}
	return err
}

