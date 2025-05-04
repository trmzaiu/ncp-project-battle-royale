# 🪖 Simple Battle Royale – Game Rules

A lightweight multiplayer Battle Royale game with simple game logic and real-time server using Go. Players join a shrinking battlefield, move, shoot, loot, and try to survive until the end!

---

## 🎮 Game Objective
Survive and become the **last player standing** in a shrinking combat zone. Eliminate opponents and avoid being caught outside the safe zone.

---

## 🧑‍🤝‍🧑 Players

- Each player starts with:
  - 100 HP
  - Empty inventory
  - A random spawn position on the map
- Players can:
  - Move (up, down, left, right)
  - Shoot other players within range
  - Pick up loot (weapons, medkits, armor)
  - View their current status (HP, position, items)

---

## 🗺️ Map & Movement

- The map is a 2D grid (e.g. 20x20).
- Each player moves one tile per action.
- The map contains loot zones and obstacles (optional).
- Map updates are broadcast to all players in real-time.

---

## 🌀 Shrinking Safe Zone ("Bo")

- The game map has a **safe zone** that **shrinks over time**.
- Players outside the zone **take damage** every few seconds.
- The safe zone shrinks in radius (circle or square) based on a timer.

**Example Shrink Rule:**
- Every 30 seconds: radius reduces by 2 tiles
- Damage outside the zone: 5 HP per second

---

## 🔫 Combat

- Players can shoot in a direction or at another player in range.
- Shooting reduces the target's HP based on weapon damage.
- When HP ≤ 0 → player is eliminated.

---

## 🧰 Loot System

- Items are randomly spawned:
  - Weapons (pistol, rifle, etc.)
  - Medkits (heal HP)
  - Armor (reduce damage taken)
- Loot is visible on the map and can be picked up when stepping on it.

---

## 🏁 Win Condition

- The game ends when:
  - Only **1 player** remains alive (Solo mode)
  - Or **1 team** remains (if Team mode is enabled)
- Winner is announced to all players.

---

## 🧪 Optional Rules / Features

- **Stamina**: Limit the number of moves per round
- **Fog of War**: Players only see nearby tiles
- **Spectator Mode**: Eliminated players can spectate live

---

## 🚀 Technologies (suggested)

- Server: Go (`net/http`, `gorilla/websocket`)
- Client (optional): HTML + JavaScript + WebSocket or Terminal UI (`tview`)
- Multiplayer: Real-time communication via WebSocket
