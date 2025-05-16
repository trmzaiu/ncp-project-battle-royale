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

	http.HandleFunc("/ws", network.HandleWebSocket)

	log.Println("Server running at http://localhost" + cfg.ServerPort)
	log.Fatal(http.ListenAndServe(cfg.ServerPort, nil))
}
