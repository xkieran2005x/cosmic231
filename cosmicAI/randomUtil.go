package cosmicAI

import (
	"math/rand"
	"time"
)

func randomRange(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - min) + min
}