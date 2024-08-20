package rpc

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ledgerwatch/log/v3"
	"golang.org/x/time/rate"
)

// RateLimitConfig contains the config of the rate limiter
type RateLimitConfig struct {
	// RateLimitApis defines the apis that need to be rate limited
	RateLimitApis []string `json:"methods"`

	// RateLimitBurst defines the maximum burst size of requests
	RateLimitCount int `json:"count"`

	// RateLimitBucket defines the time window for the rate limit
	RateLimitBucket int `json:"bucket"`
}

// RateLimit is the struct definition for the node rate limiter
type RateLimit struct {
	rlm map[string]*rate.Limiter
	sync.RWMutex
}

// gRateLimiter is the node's singleton instance for the rate limiter
var gRateLimiter = &RateLimit{
	rlm: make(map[string]*rate.Limiter),
}

// SetRateLimit sets the rate limiter singleton instance
func SetRateLimit(cfg string) {
	if cfg == "" {
		return
	}
	rlc := RateLimitConfig{}
	err := json.Unmarshal([]byte(cfg), &rlc)
	if err != nil {
		log.Warn(fmt.Sprintf("invalid rate limit config: %s", cfg))
		return
	}
	setRateLimiter(rlc)
}

// setRateLimiter sets the rate limiter in the singleton instance map
func setRateLimiter(cfg RateLimitConfig) {
	gRateLimiter.Lock()
	defer gRateLimiter.Unlock()

	// Clear rate limiter map
	gRateLimiter.rlm = make(map[string]*rate.Limiter)

	// Set API rate limiter map
	for _, api := range cfg.RateLimitApis {
		gRateLimiter.rlm[api] = rate.NewLimiter(rate.Limit(cfg.RateLimitCount), cfg.RateLimitBucket)
		log.Info(fmt.Sprintf("Rate limiter enabled for api method: %v with count: %v and bucket: %v", cfg.RateLimitApis, cfg.RateLimitCount, cfg.RateLimitBucket))
	}
	log.Info(fmt.Sprintf("Set node rate limiter, cfg: %v", cfg))
}

// checkMethodRateLimit returns true if the method API is allowed by the rate limiter
func checkMethodRateLimit(method string) bool {
	gRateLimiter.RLock()
	defer gRateLimiter.RUnlock()

	if rl, ok := gRateLimiter.rlm[method]; ok {
		return rl.Allow()
	}
	return true
}
