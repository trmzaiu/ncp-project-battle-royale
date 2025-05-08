// internal/utils/utils.go

package utils

import (
	"math/rand"
	"strconv"
	"time"
)

// IsCriticalHit checks if a critical hit occurs based on the given chance.
func IsCriticalHit(chance int) bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(100) < chance
}

func Itoa(num int) string {
	return strconv.Itoa(num)
}
