// internal/utils/loader.go

package utils

import (
	"encoding/json"
	"os"
	"royaka/internal/model"
)

func LoadTroops(path string) (map[string]*model.Troop, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]model.Troop
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	troops := make(map[string]*model.Troop)
	for name, spec := range data {
		t := spec
		t.Name = name
		troops[name] = &t
	}
	return troops, nil
}

func LoadTowers(path string) (map[string]*model.Tower, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]model.Tower
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}
	towers := make(map[string]*model.Tower)
	for name, spec := range data {
		t := spec
		towers[name] = &t
	}
	return towers, nil
}