package common

import (
	rand "math/rand/v2"
	"time"
)

type Backoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Rand      *rand.Rand
}

func NewBackoff(base, max time.Duration) Backoff {
	seed := uint64(time.Now().UnixNano())

	return Backoff{
		BaseDelay: base,
		MaxDelay:  max,

		Rand: rand.New(rand.NewPCG(seed, seed>>1)),
	}
}

func (b Backoff) Next(attempt int) time.Duration {
	d := min(b.BaseDelay<<attempt, b.MaxDelay)

	// jitter in range [0.5, 1.5)
	jitter := 0.5 + b.Rand.Float64()

	return time.Duration(float64(d) * jitter)
}
