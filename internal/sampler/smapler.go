package sampler

import (
	"fmt"
	"math/rand"
	"time"
)

type Sampler struct {
	percentage int
}

func New(rate float64) (*Sampler, error) {
	if rate <= 0 || rate > 1 {
		return nil, fmt.Errorf("incorrect sampling rate: %v", rate)
	}
	return &Sampler{
		percentage: int(rate * 100),
	}, nil
}

func (c *Sampler) IsSampled() bool {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(100)
	if randomNumber < c.percentage {
		return false
	}
	return true
}
