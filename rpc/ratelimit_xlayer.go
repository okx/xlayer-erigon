package rpc

import (
	"encoding/json"
	"sync"

	"github.com/ledgerwatch/log/v3"
	"golang.org/x/time/rate"
)

// RateLimitConfig has parameters to config the rate limit
type RateLimitConfig struct {

	// RateLimitApis defines the apis that need to be rate limited
	RateLimitApis []string `json:"methods"`

	// RateLimitBurst defines the maximum burst size of requests
	RateLimitCount int `json:"count"`

	// RateLimitBucket defines the time window for the rate limit
	RateLimitBucket int `json:"bucket"`
}

// RateLimit is the rate limit config
type RateLimit struct {
	rlm map[string]*rate.Limiter
	sync.RWMutex
}

var rateLimit = &RateLimit{}

// InitRateLimit initializes the rate limit config
func InitRateLimit(cfg string) {
	if cfg == "" {
		return
	}
	rlc := RateLimitConfig{}
	err := json.Unmarshal([]byte(cfg), &rlc)
	if err != nil {
		log.Warn("invalid rate limit config: %s", cfg)
		return
	}
	setRateLimit(rlc)
}

// setRateLimit sets the rate limit config
func setRateLimit(rlc RateLimitConfig) {
	rateLimit.Lock()
	defer rateLimit.Unlock()
	rateLimit.rlm = updateRateLimit(rlc)
}

// updateRateLimit updates the rate limit config
func updateRateLimit(rateLimit RateLimitConfig) map[string]*rate.Limiter {
	log.Info("rate limit config updated", "config", rateLimit)
	if len(rateLimit.RateLimitApis) > 0 {
		log.Info("rate limit enabled", "api", rateLimit.RateLimitApis, "count", rateLimit.RateLimitCount, "bucket", rateLimit.RateLimitBucket)
		rlm := make(map[string]*rate.Limiter)
		for _, api := range rateLimit.RateLimitApis {
			rlm[api] = rate.NewLimiter(rate.Limit(rateLimit.RateLimitCount), rateLimit.RateLimitBucket)
		}
		return rlm
	}
	return nil
}

// methodRateLimitAllow returns true if the method is allowed by the rate limit
func methodRateLimitAllow(method string) bool {
	rateLimit.RLock()
	rlm := rateLimit.rlm
	rateLimit.RUnlock()
	if rlm != nil && rlm[method] != nil && !rlm[method].Allow() {
		return false
	}
	return true
}

// ApikeyRateLimit is the api rate limit config
type ApikeyRateLimit struct {
	rlm map[string]map[string]*rate.Limiter
	sync.RWMutex
}

var apiKeyRateLimit = &ApikeyRateLimit{}

// initApikeyRateLimit initializes the apikey rate limit config
func initApikeyRateLimit(cfg map[string]*RateLimitConfig) {
	if len(cfg) == 0 {
		return
	}
	setApikeyRateLimit(cfg)
}

// setApikeyRateLimit sets the rate limit config
func setApikeyRateLimit(rlc map[string]*RateLimitConfig) {
	apiKeyRateLimit.Lock()
	defer apiKeyRateLimit.Unlock()
	apiKeyRateLimit.rlm = updateApikeyRateLimit(rlc)
}

// updateApikeyRateLimit updates the rate limit config
func updateApikeyRateLimit(rateLimit map[string]*RateLimitConfig) map[string]map[string]*rate.Limiter {
	akrlm := make(map[string]map[string]*rate.Limiter)
	log.Info("apikey rate limit config updated", "config", rateLimit)
	for apikey, config := range rateLimit {
		if len(config.RateLimitApis) > 0 {
			log.Info("rate limit enabled", "api", config.RateLimitApis, "count", config.RateLimitCount, "bucket", config.RateLimitBucket)
			rlm := make(map[string]*rate.Limiter)
			for _, api := range config.RateLimitApis {
				rlm[api] = rate.NewLimiter(rate.Limit(config.RateLimitCount), config.RateLimitBucket)
			}
			akrlm[apikey] = rlm
		}
	}
	return akrlm
}

// apikeyMethodRateLimitAllow returns true if the method is allowed by the rate limit
func apikeyMethodRateLimitAllow(api, method string) bool {
	apiKeyRateLimit.RLock()
	rlm := apiKeyRateLimit.rlm
	apiKeyRateLimit.RUnlock()
	if rlm != nil && rlm[api] != nil && rlm[api][method] != nil && !rlm[api][method].Allow() {
		return false
	}
	return true
}
