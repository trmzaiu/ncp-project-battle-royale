// internal/network/server.go

package network

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartWebSocketServer() {
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("WebSocket server listening on :8080/ws")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg.Type {
		case "login":
			var req LoginRequest
			if err := json.Unmarshal(msg.Data, &req); err != nil {
				log.Println("Invalid login data:", err)
				continue
			}
			handleLogin(conn, req)

		case "register":
			var req RegisterRequest
			if err := json.Unmarshal(msg.Data, &req); err != nil {
				log.Println("Invalid register data:", err)
				continue
			}
			handleRegister(conn, req)

		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}
}

func handleLogin(conn *websocket.Conn, req LoginRequest) {
	// Dummy check (sau này thay bằng kiểm tra thật trong file hoặc DB)
	if req.Username == "user" && req.Password == "pass" {
		resp := Response{Success: true, Message: "Login successful", Token: "dummy-token"}
		conn.WriteJSON(resp)
	} else {
		resp := Response{Success: false, Message: "Invalid credentials"}
		conn.WriteJSON(resp)
	}
}

func handleRegister(conn *websocket.Conn, req RegisterRequest) {
	// Dummy logic: luôn thành công (sau này sẽ ghi vào file JSON hoặc DB)
	log.Printf("Register user: %s\n", req.Username)

	resp := Response{Success: true, Message: "Register successful", Token: "dummy-token"}
	conn.WriteJSON(resp)
}
