package model

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ==== STRUCTS & GLOBALS ====

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
	Gold           int               `json:"gold"`
}

var (
	connToPlayer   = make(map[*websocket.Conn]*Player)
	usernameToConn = make(map[string]*websocket.Conn)
	playerData     = make(map[string]*Player)

	playerLock   sync.RWMutex
	playerDataMu sync.Mutex
)

// ==== CONSTRUCTOR ====

func NewPlayer(user *User, mode string) *Player {
	if mode != "simple" && mode != "enhanced" {
		return nil
	}

	var troops []*Troop
	var troopQueue []*Troop
	var troopInstances []*TroopInstance

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
				t := towers["King Tower"].Clone(mode, user.Level)
				t.Type = "king"
				return t
			}(),
			"guard1": func() *Tower {
				t := towers["Guard Tower"].Clone(mode, user.Level)
				t.Type = "guard1"
				return t
			}(),
			"guard2": func() *Tower {
				t := towers["Guard Tower"].Clone(mode, user.Level)
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
		Gold:           0,
	}

	return player
}

// ==== PLAYER METHODS ====

func (p *Player) RotateTroop(usedTroopName string) {
	usedIndex := -1
	for i, t := range p.Troops {
		if t.Name == usedTroopName {
			usedIndex = i
			break
		}
	}

	usedTroop := p.Troops[usedIndex]

	newTroop := p.TroopQueue[0]
	p.TroopQueue = p.TroopQueue[1:]

	p.Troops[usedIndex] = newTroop
	p.TroopQueue = append(p.TroopQueue, usedTroop)
}

func (p *Player) FullyChargeMana() {
	p.Mana = 10
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

// ==== CONNECTION MANAGEMENT ====

func RegisterConnection(conn *websocket.Conn, player *Player) {
	playerLock.Lock()
	defer playerLock.Unlock()

	connToPlayer[conn] = player
	usernameToConn[player.User.Username] = conn
	playerData[player.User.Username] = player
}

func RemoveConnection(conn *websocket.Conn) {
	playerLock.Lock()
	defer playerLock.Unlock()

	player := connToPlayer[conn]
	if player != nil {
		delete(usernameToConn, player.User.Username)
		delete(playerData, player.User.Username)
		delete(connToPlayer, conn)
	}
}

func GetPlayerByConn(conn *websocket.Conn) *Player {
	playerLock.RLock()
	defer playerLock.RUnlock()
	return connToPlayer[conn]
}

func GetConnByUsername(username string) *websocket.Conn {
	playerLock.RLock()
	defer playerLock.RUnlock()
	return usernameToConn[username]
}

func RemovePlayerByUsername(username string) {
	playerDataMu.Lock()
	defer playerDataMu.Unlock()
	delete(playerData, username)
	delete(usernameToConn, username)
}

func GetUsernameByConn(conn *websocket.Conn) string {
	playerLock.RLock()
	defer playerLock.RUnlock()

	player := connToPlayer[conn]
	if player != nil {
		return player.User.Username
	}
	return ""
}
