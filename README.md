# 🏰 Royaka

**Royaka** is a turn-based multiplayer tower battle game inspired by *Clash Royale*. Built with Go and React, it features strategic troop deployment, dynamic movement logic, and WebSocket-powered real-time gameplay.

## 🎮 Features

* 🢑 **1v1 Multiplayer Matches** via WebSocket
* ♻️ **Turn-Based Gameplay** with attack, heal, and skip mechanics
* ⚔️ **Two Game Modes**

  * **Simple Mode**: Basic strategic combat
  * **Enhanced Mode**: Adds MANA, EXP, leveling, and critical hits
* 🃏 **Troop Collection** with tanks, healers, and damage dealers
* 🧠 **Smart Troop Behavior** (e.g., river crossing only at bridges)
* 🔐 **User Authentication** (registration, login, and persistent stats)
* 📂 **File-Based Persistence** (JSON)

## 🧱 Tech Stack

* **Backend**: Go, Gorilla WebSocket, native HTTP server
* **Frontend**: React, TailwindCSS, Zustand, Lucide Icons
* **Storage**: JSON files (for users and sessions)
* **Optional**: Canvas or Pixi.js for visualizing troop movement

## 📁 Project Structure

```
royaka/
├── client/                 # React frontend (Vite)
│   ├── context/            # Zustand or React Context providers
│   ├── pages/              # Main route pages (Login, Game, etc.)
│   ├── routes/             # Route definitions and utilities
│   ├── App.jsx             # Main application layout
│   ├── main.jsx            # Entry point for React
│   └── index.html          # Vite HTML entry
├── server/                 # Go backend
│   ├── assets/             # JSON data files
│   ├── internal/
│   │ ├── game/             # Game mechanics & logic
│   │ ├── model/            # Data models (players, troops, etc.)
│   │ ├── network/          # WebSocket & HTTP handlers
│   │ └── utils/            # Utility functions
│   └── main.go             # Backend entry point
```

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/trmzaiu/royaka-clash-royale.git
cd royaka-clash-royale
```

### 2. Start the Go Backend

```bash
cd server
go mod tidy
go run main.go
```

* WebSocket endpoint: `ws://localhost:8080/ws`
* Default port: `8080`

### 3. Start the React Frontend

```bash
cd client
npm install
npm run dev
```

* Runs on `http://localhost:5173`

## 🕹️ Gameplay Overview

### 1. Turn-Based Mode
- Players take turns (Player 1 → Player 2).
- Each turn, the player gains 3 mana.
- If a player does not have enough mana to attack, they can choose to skip their turn.
- Each player has 30 seconds per turn; if no action is taken, the turn automatically passes to the opponent.
- Victory requires destroying both Guard Towers before accessing the King Tower.

### 2. Timed Match Mode
- Mana increases automatically over time (1 mana every 2 seconds).
- Both players act simultaneously in real-time.
- Towers actively defend by attacking enemy troops within range.
- Matches last 3 minutes with fast-paced, continuous action.
- Victory conditions remain the same: eliminate both Guard Towers before accessing the King Tower.

## 🔐 Authentication System

* Users register and log in via HTTP
* Passwords hashed using bcrypt
* Sessions are stored in a `sessions.json` file
* Match stats (wins/losses) saved per user

---

Made with ❤️ by game dev enthusiasts.
