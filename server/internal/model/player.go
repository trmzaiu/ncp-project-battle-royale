// internal/model/player.go

package model

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	Mana           int               `json:"mana"`
	Towers         map[string]*Tower `json:"towers"`
	TowerInstances []*TowerInstance  `json:"tower_instances,omitempty"`
	Troops         []*Troop          `json:"troops"`
	TroopQueue     []*Troop          `json:"-"`
	TroopInstances []*TroopInstance  `json:"troop_instances,omitempty"`
	Active         bool              `json:"active"`
	User           *User             `json:"user"`
	Matched        chan bool         `json:"-"`
	Turn           int               `json:"turn"`
	LastManaRegen  time.Time         `json:"-"`
}

var (
	connToPlayer   = make(map[*websocket.Conn]*Player)
	usernameToConn = make(map[string]*websocket.Conn)
	playerLock     sync.RWMutex
	playerData     = make(map[string]*Player)
	playerDataMu   sync.Mutex
)

func NewPlayer(user *User, mode string) *Player {
	if mode != "simple" && mode != "enhanced" {
		return nil
	}

	var (
		troops         []*Troop
		troopQueue     []*Troop
		troopInstances []*TroopInstance
	)

	if mode == "simple" {
		troops = getRandomTroops(4)
	} else {
		allTroops := getRandomTroops(8)
		shuffled := shuffleTroops(allTroops)
		troops = shuffled[:4]
		troopQueue = shuffled[4:]
		troopInstances = createTroopInstances(troops, user.Username)
	}

	towers := LoadTower()

	player := &Player{
		Mana: 5,
		Towers: map[string]*Tower{
			"king": func() *Tower {
				t := towers["King Tower"].Clone()
				t.Type = "king"
				return t
			}(),
			"guard1": func() *Tower {
				t := towers["Guard Tower"].Clone()
				t.Type = "guard1"
				return t
			}(),
			"guard2": func() *Tower {
				t := towers["Guard Tower"].Clone()
				t.Type = "guard2"
				return t
			}(),
		},
		TowerInstances: []*TowerInstance{},
		Troops:         troops,
		TroopQueue:     troopQueue,
		TroopInstances: troopInstances,
		Active:         true,
		User:           user,
		Matched:        make(chan bool, 1),
		Turn:           0,
		LastManaRegen:  time.Now(),
	}

	return player
}

func (p *Player) RotateTroop(usedTroopName string) {
	usedIndex := -1
	for i, t := range p.Troops {
		if t.Name == usedTroopName {
			usedIndex = i
			break
		}
	}
	if usedIndex == -1 || p.Mana < p.Troops[usedIndex].MANA || len(p.TroopQueue) == 0 {
		return
	}

	usedTroop := p.Troops[usedIndex]

	p.Mana -= usedTroop.MANA

	if len(p.TroopQueue) == 0 {
		return
	}

	newTroop := p.TroopQueue[0]
	p.TroopQueue = p.TroopQueue[1:]

	p.Troops[usedIndex] = newTroop

	p.TroopQueue = append(p.TroopQueue, usedTroop)
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

func (p *Player) Reset(mode string) {
	p.Mana = 5
	p.Towers["king"].Reset("king")
	p.Towers["guard1"].Reset("guard1")
	p.Towers["guard2"].Reset("guard2")

	if mode == "simple" {
		p.Troops = getRandomTroops(4)
		p.TroopInstances = nil
	} else {
		p.Troops = getRandomTroops(8)
		p.TroopInstances = createTroopInstances(p.Troops, p.User.ID)
	}

	p.Active = false
	p.Turn = 0
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
