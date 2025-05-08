package player

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrUserExists      = errors.New("user already exists")
	ErrUserNotFound    = errors.New("user not found")
	playersFile        = "assets/data/players.json"
	playersStorageLock = &sync.Mutex{}
)

// InitStorage ensures the data directory exists
func InitStorage() error {
	dir := filepath.Dir(playersFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(playersFile); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		return ioutil.WriteFile(playersFile, []byte("[]"), 0644)
	}
	return nil
}

// LoadPlayers loads all players from storage
func LoadPlayers() ([]Player, error) {
	playersStorageLock.Lock()
	defer playersStorageLock.Unlock()

	if err := InitStorage(); err != nil {
		return nil, err
	}

	file, err := os.Open(playersFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var players []Player
	err = json.NewDecoder(file).Decode(&players)
	return players, err
}

// SavePlayers persists players to storage
func SavePlayers(players []Player) error {
	playersStorageLock.Lock()
	defer playersStorageLock.Unlock()

	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(playersFile, data, 0644)
}

// AddPlayer adds a new player if username is unique
func AddPlayer(newPlayer Player) error {
	players, err := LoadPlayers()
	if err != nil {
		return err
	}

	// Check if username already exists
	for _, p := range players {
		if p.Username == newPlayer.Username {
			return ErrUserExists
		}
	}

	players = append(players, newPlayer)
	return SavePlayers(players)
}

// FindPlayerByUsername retrieves a player by username
func FindPlayerByUsername(username string) (bool) {
	players, err := LoadPlayers()
	if err != nil {
		return false
	}

	for _, p := range players {
		if p.Username == username {
			return true
		}
	}
	return false
}