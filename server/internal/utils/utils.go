// internal/utils/utils.go

package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func Itoa(num int) string {
	return strconv.Itoa(num)
}

func GenerateRoomID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func AbsFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func ClampFloat(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func CryptoRandInt(max int64) (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}