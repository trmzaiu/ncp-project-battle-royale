package main

var arenaRadius = 100 // Initial radius of the arena

// Shrink the arena over time
func shrinkArena() {
    if arenaRadius > 10 {
        arenaRadius -= 5
    }
}

// Check if a player is within the arena boundaries
func isPlayerInArena(player *Player) bool {
    return player.X >= -arenaRadius && player.X <= arenaRadius && player.Y >= -arenaRadius && player.Y <= arenaRadius
}
