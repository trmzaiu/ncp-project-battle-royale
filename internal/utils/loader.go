// internal/utils/loader.go

package utils

import (
	"encoding/json"
	"os"
	"royaka/internal/player"
)

func LoadTroops(path string) (map[string]*player.Troop, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]player.Troop
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	troops := make(map[string]*player.Troop)
	for name, spec := range data {
		t := spec
		t.Name = name
		troops[name] = &t
	}
	return troops, nil
}

func LoadTowers(path string) (map[string]*player.Tower, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]player.Tower
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}
	towers := make(map[string]*player.Tower)
	for name, spec := range data {
		t := spec
		towers[name] = &t
	}
	return towers, nil
}