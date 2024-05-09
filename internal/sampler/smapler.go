package sampler

import (
	"math/rand"
	"time"
)

type Sampler struct {
	percentage int
}

func New(rate float64) *Sampler {
	return &Sampler{
		percentage: int(rate * 100),
	}
}

func (c *Sampler) IsSampled() bool {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(100)
	if randomNumber < c.percentage {
		return false
	}
	return true
}
