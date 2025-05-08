// internal/player/player.go

package player

type Player struct {
	Username string `json:"username"`
	Password string `json:"password"` // In production: hash this!
}
