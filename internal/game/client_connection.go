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

    return c.Conn.WriteJSON(data)
}

