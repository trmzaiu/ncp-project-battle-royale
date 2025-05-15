// internal/model/player.go

package model

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	Mana    int               `json:"mana"`
	Towers  map[string]*Tower `json:"towers"`
	Troops  []*Troop          `json:"troops"`
	Active  bool              `json:"active"`
	User    *User             `json:"user"`
	Matched chan bool         `json:"-"`
}

var (
	connToPlayer   = make(map[*websocket.Conn]*Player)
	usernameToConn = make(map[string]*websocket.Conn) // Đảm bảo lấy được kết nối từ username
	playerLock     sync.RWMutex
	playerData     = make(map[string]*Player) // username -> player
	playerDataMu   sync.Mutex
)

func NewPlayer(user *User, mode string) *Player {
	var troops []*Troop
	if mode != "simple" && mode != "enhanced" {
		return nil
	} else if mode == "simple" {
		troops = getRandomTroops(4)
	} else {
		troops = getRandomTroops(6)
	}

	towers := LoadTower()

	return &Player{
		Mana: 5,
		Towers: map[string]*Tower{
			"king":   towers["King Tower"].Clone(),
			"guard1": towers["Guard Tower"].Clone(),
			"guard2": towers["Guard Tower"].Clone(),
		},
		Troops:  troops,
		Active:  true,
		User:    user,
		Matched: make(chan bool, 1),
	}
}

func (p *Player) ApplyDefenseBoost(percent float64) {
	for _, t := range p.Towers {
		t.IncreaseDefense(percent)
	}
}

func (p *Player) BoostAllTroops() {
	for _, t := range p.Troops {
		t.BoostAttack()
	}
}

func (p *Player) FullyChargeMana() {
	p.Mana = 10
}

func (p *Player) Reset() {
	p.Mana = 5
	p.Towers["king"].Reset("king")
	p.Towers["guard1"].Reset("guard1")
	p.Towers["guard2"].Reset("guard2")
	p.Troops = getRandomTroops(4)
	p.Active = false
}

func (p *Player) DestroyedCount() int {
	count := 0
	for _, t := range p.Towers {
		if t.HP <= 0 {
			count++
		}
	}
	return count
}

// Register a new player connection
func RegisterConnection(conn *websocket.Conn, player *Player) {
	playerLock.Lock()
	defer playerLock.Unlock()

	connToPlayer[conn] = player
	usernameToConn[player.User.Username] = conn // Mapping username -> conn
	playerData[player.User.Username] = player   // Mapping username -> player
}

// Remove connection mapping when player disconnects
func RemoveConnection(conn *websocket.Conn) {
	playerLock.Lock()
	defer playerLock.Unlock()

	player := connToPlayer[conn]
	if player != nil {
		delete(usernameToConn, player.User.Username) // Remove username -> conn mapping
		delete(playerData, player.User.Username)     // Remove username -> player mapping
		delete(connToPlayer, conn)                   // Remove conn -> player mapping
	}
}

// Get player by connection
func GetPlayerByConn(conn *websocket.Conn) *Player {
	playerLock.RLock()
	defer playerLock.RUnlock()
	return connToPlayer[conn]
}

// Get connection by username
func GetConnByUsername(username string) *websocket.Conn {
	playerLock.RLock()
	defer playerLock.RUnlock()
	return usernameToConn[username]
}

// Remove player data by username
func RemovePlayerByUsername(username string) {
	playerDataMu.Lock()
	defer playerDataMu.Unlock()
	delete(playerData, username)
	delete(usernameToConn, username)
}

func GetUsernameByConn(conn *websocket.Conn) string {
	playerLock.RLock()
	defer playerLock.RUnlock()

	// Truy tìm player từ conn
	player := connToPlayer[conn]
	if player != nil {
		return player.User.Username
	}
	return ""
}
