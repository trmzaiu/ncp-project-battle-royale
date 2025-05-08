// internal/player/store.go

package player

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var playersFile = "assets/data/players.json"

func LoadPlayers() ([]Player, error) {
	if _, err := os.Stat(playersFile); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		err := ioutil.WriteFile(playersFile, []byte("[]"), 0644)
		if err != nil {
			return nil, err
		}
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


func SavePlayers(players []Player) error {
	data, err := json.MarshalIndent(players, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(playersFile, data, 0644)
}

func AddPlayer(newPlayer Player) error {
	players, err := LoadPlayers()
	if err != nil {
		return err
	}
	players = append(players, newPlayer)
	return SavePlayers(players)
}

func FindPlayer(username, password string) bool {
	players, err := LoadPlayers()
	if err != nil {
		return false
	}
	for _, p := range players {
		if p.Username == username && p.Password == password {
			return true
		}
	}
	return false
}
