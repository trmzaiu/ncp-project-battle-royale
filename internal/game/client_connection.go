package game

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ClientConnection struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
	IsClosed  bool
    Username  string 
}

func (c *ClientConnection) SafeWrite(data interface{}) error {
    c.Mu.Lock()
    defer c.Mu.Unlock()

    if c.Conn == nil {
        log.Println("[WS] No connection to write to")
        return nil
    }

    c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
    defer c.Conn.SetWriteDeadline(time.Time{})

    err := c.Conn.WriteJSON(data)
    if err != nil {
        log.Printf("[WS] Failed to send: %v", err)
        if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
            log.Printf("[WS] WebSocket unexpectedly closed: %v", err)
            c.Conn = nil
        }
    } else {
        log.Printf("[WS] Sent: %T", data)
    }
    return err
}

