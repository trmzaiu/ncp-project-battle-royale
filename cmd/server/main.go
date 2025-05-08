// cmd/server/main.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"royaka/internal/network"
)

var users = make(map[string]string) // In-memory store for simplicity

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./cmd/client/static"))))

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req network.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check user credentials
	storedPassword, exists := users[req.Username]
	if !exists || storedPassword != req.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Send response
	response := network.Response{
		Success: true,
		Message: "Login successful",
	}
	json.NewEncoder(w).Encode(response)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req network.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	if _, exists := users[req.Username]; exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Register user
	users[req.Username] = req.Password

	// Send response
	response := network.Response{
		Success: true,
		Message: "Registration successful",
	}
	json.NewEncoder(w).Encode(response)
}



