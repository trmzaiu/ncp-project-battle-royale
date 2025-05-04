package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    // Set up WebSocket route
    http.HandleFunc("/ws", handleWebSocket)
    fmt.Println("🔫 Game server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
