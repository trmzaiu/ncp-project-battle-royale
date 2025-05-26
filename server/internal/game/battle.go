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

func (bs *BattleSystem) AddEntity(e BattleEntity) {
	bs.MapMutex.Lock()
	defer bs.MapMutex.Unlock()
	pos := e.GetPosition().String()
	bs.BattleMap[pos] = append(bs.BattleMap[pos], e)
}

func (bs *BattleSystem) GetEntities() []BattleEntity {
	bs.MapMutex.RLock()
	defer bs.MapMutex.RUnlock()
	var result []BattleEntity
	for _, list := range bs.BattleMap {
		result = append(result, list...)
	}
	return result
}

func (bs *BattleSystem) GetEntityList() []BattleEntity {
	return bs.GetEntities() // alias
}

func (bs *BattleSystem) CleanupDeadEntities() {
	bs.MapMutex.Lock()
	defer bs.MapMutex.Unlock()

	for key, list := range bs.BattleMap {
		live := list[:0]
		for _, e := range list {
			if e.IsAlive() {
				live = append(live, e)
			}
		}
		if len(live) == 0 {
			delete(bs.BattleMap, key)
		} else {
			bs.BattleMap[key] = live
		}
	}
}

func (bs *BattleSystem) Stop() {
	select {
	case bs.TickerStopChan <- struct{}{}:
	default:
	}
}
