// cmd/server/main.go

package main

import (
	"log"
	"net/http"
	"royaka/config"
	"royaka/internal/network"
)

func main() {
	cfg := config.LoadConfig()

	// Serve static HTML
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	

	// WebSocket endpoint
	http.HandleFunc("/ws", network.HandleWS)

	log.Println("Server running at http://localhost" + cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, nil))
}
