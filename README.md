# 🏰 Royaka

**Royaka** is a turn-based multiplayer tower battle game inspired by *Clash Royale*. Built with Go and React, it features strategic troop deployment, dynamic movement logic, and WebSocket-powered real-time gameplay.

## 🎮 Features

- 🧑‍🤝‍🧑 **1v1 Multiplayer Matches** via WebSocket
- 🔁 **Turn-Based Gameplay** with attack, heal, and skip mechanics
- ⚔️ **Two Game Modes**
  - **Simple Mode**: Basic strategic combat
  - **Enhanced Mode**: Adds MANA, EXP, leveling, and critical hits
- 🃏 **Troop Collection** with tanks, healers, and damage dealers
- 🧠 **Smart Troop Behavior** (e.g., river crossing only at bridges)
- 🔐 **User Authentication** (registration, login, and persistent stats)
- 💾 **File-Based Persistence** (JSON)

## 🧱 Tech Stack

- **Backend**: Go, Gorilla WebSocket, native HTTP server
- **Frontend**: React, TailwindCSS, Zustand, Lucide Icons
- **Storage**: JSON files (for users and sessions)
- **Optional**: Canvas or Pixi.js for visualizing troop movement

## 📁 Project Structure
<pre> ## 📁 Project Structure ``` royaka/ ├── client/ # React frontend │ ├── components/ # UI components │ ├── pages/ # Game routes │ └── main.tsx # App entry ├── server/ # Go backend │ ├── game/ # Game mechanics & logic │ ├── handlers/ # WebSocket & HTTP handlers │ ├── models/ # Structs for players, troops, etc. │ └── main.go # Backend entry point ``` </pre>



