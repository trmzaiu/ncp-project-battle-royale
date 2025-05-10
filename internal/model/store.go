// internal/model/store.go

package model

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
	usersFile          = "assets/data/users.json"
	usersStorageLock   = &sync.Mutex{}
)

// InitStorage ensures the data directory exists
func InitStorage() error {
	dir := filepath.Dir(usersFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		return ioutil.WriteFile(usersFile, []byte("[]"), 0644)
	}
	return nil
}

// LoadUsers loads all users from storage
func LoadUsers() ([]User, error) {
	usersStorageLock.Lock()
	defer usersStorageLock.Unlock()

	if err := InitStorage(); err != nil {
		return nil, err
	}

	file, err := os.Open(usersFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users []User
	err = json.NewDecoder(file).Decode(&users)
	return users, err
}

// SaveUsers persists users to storage
func SaveUsers(users []User) error {
	usersStorageLock.Lock()
	defer usersStorageLock.Unlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(usersFile, data, 0644)
}

func SaveUser(user User) error {
	users, err := LoadUsers()
	if err != nil {
		return err
	}

	updated := false
	for i, u := range users {
		if u.Username == user.Username {
			users[i] = user
			updated = true
			break
		}
	}

	if !updated {
		users = append(users, user)
	}

	return SaveUsers(users)
}

// AddUser adds a new user if the username is unique
func AddUser(newUser User) error {
	users, err := LoadUsers()
	if err != nil {
		return err
	}

	// Check if the username already exists
	for _, u := range users {
		if u.Username == newUser.Username {
			return ErrUserExists
		}
	}

	users = append(users, newUser)
	return SaveUsers(users)
}

// FindUserByUsername retrieves a user by username
func FindUserByUsername(username string) (User, bool) {
    users, err := LoadUsers()
    if err != nil {
        return User{}, false
    }

    for _, u := range users {
        if u.Username == username {
            return u, true
        }
    }
    return User{}, false
}
