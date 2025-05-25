// internal/model/user.go

package model

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"
)

// ==== STRUCTS ====

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"` 
	CreatedAt   time.Time `json:"createdAt"`
	LastLogin   time.Time `json:"lastLogin"`
	IsActive    bool      `json:"isActive"`
	EXP         int       `json:"exp"`
	Level       int       `json:"level"`
	GamesPlayed int       `json:"gamesPlayed"` // Track number of games played
	GamesWon    int       `json:"gamesWon"`    // Track number of games won
	Avatar      string    `json:"avatar"`
	Gold        int       `json:"gold"`
}

func NewUser(username, password string) *User {
	avatar := getRandomAvatar()

	return &User{
		ID:          generateID(),
		Username:    username,
		Password:    password, // Should be hashed in actual implementation
		CreatedAt:   time.Now(),
		LastLogin:   time.Now(),
		IsActive:    true,
		EXP:         0,
		Level:       1,
		GamesPlayed: 0,
		GamesWon:    0,
		Avatar:      strconv.Itoa(avatar),
	}
}

// Helper function for ID generation
func generateID() string {
	timestamp := time.Now().Format("20060102150405")
	randomPart := randomString(8)
	return timestamp + randomPart
}

// Helper function for random string generation
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic("failed to generate secure random string: " + err.Error())
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func (u *User) AddExp(amount int) {
	u.EXP += amount

	for {
		maxExp := GetMaxExp(u.Level)
		if u.EXP >= maxExp {
			u.EXP -= maxExp
			u.Level++
		} else {
			break
		}
	}
}

type Level struct {
	Level  int `json:"level"`
	MaxExp int `json:"max_exp"`
}

func GetMaxExp(level int) int {
	if level <= 0 {
		return 0
	}

	exp := 100.0
	for i := 1; i < level; i++ {
		exp *= 1.5
	}
	return int(exp)
}

func GenerateLevels(maxLevel int) []Level {
	levels := make([]Level, maxLevel)
	exp := 100.0

	for i := 0; i < maxLevel; i++ {
		levels[i] = Level{
			Level:  i + 1,
			MaxExp: int(exp),
		}
		exp *= 1.5
	}

	return levels
}

func getRandomAvatar() int {
	n, err := rand.Int(rand.Reader, big.NewInt(15)) // 0–14
	if err != nil {
		// fallback or panic, your call
		return 1
	}
	return int(n.Int64()) + 1 // shift to 1–15
}
