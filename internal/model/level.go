// internal/model/level.go

package model

type Level struct {
	Level   int `json:"level"`
	MaxExp  int `json:"max_exp"`
}

// Example level system with increasing EXP requirements
var Levels = []Level{
	{Level: 1, MaxExp: 100},
	{Level: 2, MaxExp: 200},
	{Level: 3, MaxExp: 350},
	{Level: 4, MaxExp: 500},
	{Level: 5, MaxExp: 700},
	{Level: 6, MaxExp: 1000},
	{Level: 7, MaxExp: 1350},
	{Level: 8, MaxExp: 1750},
	{Level: 9, MaxExp: 2200},
	{Level: 10, MaxExp: 2700},
}

// GetMaxExp returns the max EXP required for a given level
func GetMaxExp(level int) int {
	if level <= 0 {
		return 0
	}
	if level > len(Levels) {
		lastLevel := Levels[len(Levels)-1]
		return lastLevel.MaxExp + 500*(level-len(Levels))
	}
	return Levels[level-1].MaxExp
}
