package rpc

import (
	"strings"
	"sync"

	"github.com/ledgerwatch/log/v3"
	"github.com/spf13/cast"
	"golang.org/x/time/rate"
)

// RateLimitConfig has parameters to config the rate limit
type RateLimitConfig struct {

	// RateLimitApis defines the apis that need to be rate limited
	RateLimitApis []string `mapstructure:"RateLimitApis"`

	// RateLimitBurst defines the maximum burst size of requests
	RateLimitCount int `mapstructure:"RateLimitCount"`

	// RateLimitDuration defines the time window for the rate limit
	RateLimitDuration int `mapstructure:"RateLimitDuration"`
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
	parties := strings.Split(cfg, ":")
	if len(parties) != 3 {
		log.Warn("invalid rate limit config: %s", cfg)
		return
	}
	rlc := RateLimitConfig{}
	rlc.RateLimitApis = strings.Split(parties[0], "|")
	if len(rlc.RateLimitApis) == 0 {
		log.Warn("invalid rate limit apis: %s", parties[0])
		return
	}
	rlc.RateLimitCount = cast.ToInt(parties[1])
	rlc.RateLimitDuration = cast.ToInt(parties[2])

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
		log.Info("rate limit enabled", "api", rateLimit.RateLimitApis, "count", rateLimit.RateLimitCount, "duration", rateLimit.RateLimitDuration)
		rlm := make(map[string]*rate.Limiter)
		for _, api := range rateLimit.RateLimitApis {
			rlm[api] = rate.NewLimiter(rate.Limit(rateLimit.RateLimitCount), rateLimit.RateLimitDuration)
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
