// internal/model/user.go

package model

import "time"

type User struct {
    ID          string    `json:"id"`
    Username    string    `json:"username"`
    Password    string    `json:"password"` // This should be hashed in production
    CreatedAt   time.Time `json:"createdAt"`
    LastLogin   time.Time `json:"lastLogin"`
    IsActive    bool      `json:"isActive"`
    EXP         int       `json:"exp"`
    Level       int       `json:"level"`
    GamesPlayed int       `json:"gamesPlayed"` // Track number of games played
    GamesWon    int       `json:"gamesWon"`    // Track number of games won
}

func NewUser(username, password string) *User {
	return &User{
		ID:        generateID(), // Implement your ID generation function
		Username:  username,
		Password:  password, // Should be hashed in actual implementation
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
		IsActive:  true,
		EXP:       0,
		Level:     1,
		GamesPlayed: 0,
		GamesWon:    0,
	}
}

// Helper function for ID generation
func generateID() string {
	// In a real implementation, use a more robust ID generation method
	return time.Now().Format("20060102150405") + randomString(8)
}

// Helper function for random string generation
func randomString(length int) string {
	// Implement a proper random string generator
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[i%len(chars)]
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