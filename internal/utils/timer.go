// internal/utils/timer.go

package utils

import (
	"sync"
	"time"
)

// GameTimer represents a countdown timer for match duration
type GameTimer struct {
	duration  time.Duration
	startTime time.Time
	active    bool
	mu        sync.RWMutex
}

// NewGameTimer creates a new game timer with the given duration in seconds
func NewGameTimer(durationSecs int) *GameTimer {
	return &GameTimer{
		duration: time.Duration(durationSecs) * time.Second,
		active:   false,
	}
}

// Start begins the timer countdown
func (t *GameTimer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.startTime = time.Now()
	t.active = true
}

// Stop stops the timer
func (t *GameTimer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.active = false
}

// Reset resets the timer to its initial state
func (t *GameTimer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.active = false
}

// TimeRemaining returns the remaining time in seconds
func (t *GameTimer) TimeRemaining() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	if !t.active {
		return int(t.duration.Seconds())
	}
	
	elapsed := time.Since(t.startTime)
	remaining := t.duration - elapsed
	
	if remaining < 0 {
		return 0
	}
	
	return int(remaining.Seconds())
}

// IsExpired checks if the timer has expired
func (t *GameTimer) IsExpired() bool {
	return t.TimeRemaining() <= 0
}

// ManaTimer represents a timer for mana regeneration
type ManaTimer struct {
	lastRegenTime time.Time
	regenRate     float64 // seconds per mana point
	mu            sync.RWMutex
}

// NewManaTimer creates a new mana timer with the given regeneration rate
func NewManaTimer(regenRate float64) *ManaTimer {
	return &ManaTimer{
		lastRegenTime: time.Now(),
		regenRate:     regenRate,
	}
}

// Reset resets the mana timer
func (t *ManaTimer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRegenTime = time.Now()
}

// ManaGained calculates how much mana has been gained since the last call
// and updates the last regeneration time
func (t *ManaTimer) ManaGained() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(t.lastRegenTime).Seconds()
	
	// Calculate how many mana points have been regenerated
	manaGained := int(elapsed / t.regenRate)
	
	// Update the last regeneration time, accounting for any partial mana
	if manaGained > 0 {
		t.lastRegenTime = t.lastRegenTime.Add(time.Duration(float64(manaGained) * t.regenRate * float64(time.Second)))
	}
	
	return manaGained
}