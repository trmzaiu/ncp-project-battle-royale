package main

import (
	"log"
	"net/http"
	"os"
	"royaka/internal/network"
)

func main() {
	// cfg := config.LoadConfig()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// WebSocket handler
	http.HandleFunc("/ws", network.HandleWebSocket)

	log.Println("Server running at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
