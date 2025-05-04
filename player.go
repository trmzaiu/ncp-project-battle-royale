package main

import "github.com/gorilla/websocket"

type Player struct {
    ID    string
    X, Y  int
    Alive bool
    Conn  *websocket.Conn // WebSocket connection
}

// Store players in a global map
var players = make(map[string]*Player)

// Add a new player to the game
func addPlayer(id string, conn *websocket.Conn) *Player {
    player := &Player{
        ID:    id,
        X:     0,
        Y:     0,
        Alive: true,
        Conn:  conn,
    }
    players[id] = player
    return player
}

// Move the player in the specified direction
func movePlayer(p *Player, direction string) {
    switch direction {
    case "up":
        p.Y--
    case "down":
        p.Y++
    case "left":
        p.X--
    case "right":
        p.X++
    }
}
