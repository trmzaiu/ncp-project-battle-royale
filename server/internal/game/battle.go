package game

import (
	"royaka/internal/model"
	"sync"
	"time"
)

type BattleSystem struct {
	BattleMap      map[string][]BattleEntity
	MapMutex       sync.RWMutex
	TickerStopChan chan struct{}
	TickRate       time.Duration
}

type BattleEntity interface {
	GetID() string
	GetOwner() string
	GetType() string
	GetPosition() model.Position
	IsAlive() bool
}

func NewBattleSystem(tickRate time.Duration) *BattleSystem {
	return &BattleSystem{
		BattleMap:      make(map[string][]BattleEntity),
		TickerStopChan: make(chan struct{}),
		TickRate:       tickRate,
	}
}

// AddEntity safely adds a new entity
func (bs *BattleSystem) AddEntity(entity BattleEntity) {
	bs.MapMutex.Lock()
	defer bs.MapMutex.Unlock()
	positionKey := entity.GetPosition().String() // Make sure Position.String() returns "x_y" format
	bs.BattleMap[positionKey] = append(bs.BattleMap[positionKey], entity)
}

// GetEntities safely returns a copy of all entities
func (bs *BattleSystem) GetEntities() []BattleEntity {
    bs.MapMutex.RLock()
    defer bs.MapMutex.RUnlock()

    var entities []BattleEntity
    for _, list := range bs.BattleMap {
        // log.Printf("[DEBUG] Pos %s has %d entities", k, len(list))
        entities = append(entities, list...)
    }
    return entities
}

func (bs *BattleSystem) GetEntityList() []BattleEntity {
    bs.MapMutex.RLock()
    defer bs.MapMutex.RUnlock()

    var entities []BattleEntity
    for _, list := range bs.BattleMap {
        entities = append(entities, list...)
    }
    return entities
}

func (bs *BattleSystem) CleanupDeadEntities() {
	bs.MapMutex.Lock()
	defer bs.MapMutex.Unlock()

	for posKey, ents := range bs.BattleMap {
		alive := ents[:0] // reuse underlying array
		for _, e := range ents {
			if e.IsAlive() {
				alive = append(alive, e)
			}
		}
		if len(alive) == 0 {
			delete(bs.BattleMap, posKey)
		} else {
			bs.BattleMap[posKey] = alive
		}
	}
}

// Stop safely stops the battle system
func (bs *BattleSystem) Stop() {
	select {
	case bs.TickerStopChan <- struct{}{}:
		// Stop signal sent
	default:
		// Already stopping or stopped
	}
}
