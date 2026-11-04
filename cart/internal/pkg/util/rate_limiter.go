package util

import (
	"math"
	"sync"
	"time"
)

// RateLimiter is used to control the rate of requests.
type RateLimiter struct {
	mutex sync.Mutex

	tokens    int
	maxTokens int

	interval   time.Duration
	lastRefill time.Time
}

// NewRateLimiter constructor for RateLimiter.
func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	if maxTokens < 1 {
		maxTokens = math.MaxInt
	}

	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		interval:   interval / time.Duration(maxTokens),
		lastRefill: time.Now(),
	}
}

// Allow checks the ability to perform some action.
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	diff := now.Sub(rl.lastRefill)
	rl.tokens += int(diff / rl.interval)
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// Wait blocks the execution until RateLimiter allows to continue.
func (rl *RateLimiter) Wait() {
	for !rl.Allow() {
		time.Sleep(rl.interval)
	}
}
