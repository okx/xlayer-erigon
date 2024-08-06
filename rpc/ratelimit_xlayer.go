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

// InitRateLimit initializes the rate limiter singleton instance
func InitRateLimit(cfg string) {
	if cfg == "" {
		return
	}
	rlc := RateLimitConfig{}
	err := json.Unmarshal([]byte(cfg), &rlc)
	if err != nil {
		log.Warn(fmt.Sprintf("invalid rate limit config: %s", cfg))
		return
	}
	SetRateLimit(rlc)
}

// SetRateLimit sets the rate limiter singleton instance
func SetRateLimit(cfg RateLimitConfig) {
	gRateLimiter.Lock()
	defer gRateLimiter.Unlock()

	log.Info(fmt.Sprintf("Setting node rate limiter, cfg: %v", cfg))
	for _, api := range cfg.RateLimitApis {
		gRateLimiter.rlm[api] = rate.NewLimiter(rate.Limit(cfg.RateLimitCount), cfg.RateLimitBucket)
		log.Info(fmt.Sprintf("Rate limiter enabled for api method: %v with count: %v and bucket: %v", cfg.RateLimitApis, cfg.RateLimitCount, cfg.RateLimitBucket))
	}
}

// ApikeyRateLimit is the struct definition for the API key rate limiter
type ApikeyRateLimit struct {
	rlm map[string]map[string]*rate.Limiter
	sync.RWMutex
}

// gApikeyRateLimiter is the node's singleton instance for the API key rate limiter
var gApikeyRateLimiter = &ApikeyRateLimit{
	rlm: make(map[string]map[string]*rate.Limiter),
}

// setApiKeyRateLimit sets the global API key rate limiter
func setApikeyRateLimit(key string, cfg RateLimitConfig) {
	gApikeyRateLimiter.Lock()
	defer gApikeyRateLimiter.Unlock()

	if _, ok := gApikeyRateLimiter.rlm[key]; ok {
		log.Warn("API key rate limiter already set, skipping")
		return
	}

	log.Info(fmt.Sprintf("Setting API key rate limiter for key: %v, cfg: %v", key, cfg))
	gApikeyRateLimiter.rlm[key] = make(map[string]*rate.Limiter)
	for _, api := range cfg.RateLimitApis {
		gApikeyRateLimiter.rlm[key][api] = rate.NewLimiter(rate.Limit(cfg.RateLimitCount), cfg.RateLimitBucket)
		log.Info(fmt.Sprintf("Rate limiter enabled for key: %v for api method: %v with count: %v and bucket: %v", key, cfg.RateLimitApis, cfg.RateLimitCount, cfg.RateLimitBucket))
	}
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

// checkApikeyMethodRateLimit returns true if the key and the method API is allowed
// by the API key rate limiter
func checkApikeyMethodRateLimit(key, method string) bool {
	gApikeyRateLimiter.RLock()
	defer gApikeyRateLimiter.RUnlock()

	if rlm, keyFound := gApikeyRateLimiter.rlm[key]; keyFound {
		if rl, ok := rlm[method]; ok {
			return rl.Allow()
		}
	}
	return true
}
