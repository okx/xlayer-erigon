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

// ApikeyRateLimit is the api rate limit config
type ApikeyRateLimit struct {
	rlm map[string]map[string]*rate.Limiter
	sync.RWMutex
}

var apiKeyRateLimit = &ApikeyRateLimit{}

// initApikeyRateLimit initializes the apikey rate limit config
func initApikeyRateLimit(cfg map[string]string) {
	if len(cfg) == 0 {
		return
	}
	var mconfig = make(map[string]RateLimitConfig)
	for apikey, config := range cfg {
		parties := strings.Split(config, ":")
		if len(parties) != 3 {
			log.Warn("invalid rate limit config: %s", config)
			continue
		}
		rlc := RateLimitConfig{}
		rlc.RateLimitApis = strings.Split(parties[0], "|")
		if len(rlc.RateLimitApis) == 0 {
			log.Warn("invalid rate limit apis: %s", parties[0])
			continue
		}
		count, err := cast.ToIntE(parties[1])
		if err != nil {
			log.Warn("invalid rate limit count: %s", parties[1])
			continue
		}
		duration, err := cast.ToIntE(parties[2])
		if err != nil {
			log.Warn("invalid rate limit duration: %s", parties[2])
			continue
		}
		rlc.RateLimitCount = count
		rlc.RateLimitDuration = duration
		mconfig[apikey] = rlc
	}
	setApikeyRateLimit(mconfig)
}

// setApikeyRateLimit sets the rate limit config
func setApikeyRateLimit(rlc map[string]RateLimitConfig) {
	apiKeyRateLimit.Lock()
	defer apiKeyRateLimit.Unlock()
	apiKeyRateLimit.rlm = updateApikeyRateLimit(rlc)
}

// updateApikeyRateLimit updates the rate limit config
func updateApikeyRateLimit(rateLimit map[string]RateLimitConfig) map[string]map[string]*rate.Limiter {
	log.Info("apikey rate limit config updated", "config", rateLimit)
	akrlm := make(map[string]map[string]*rate.Limiter)
	for apikey, config := range rateLimit {
		if len(config.RateLimitApis) > 0 {
			log.Info("rate limit enabled", "api", config.RateLimitApis, "count", config.RateLimitCount, "duration", config.RateLimitDuration)
			rlm := make(map[string]*rate.Limiter)
			for _, api := range config.RateLimitApis {
				rlm[api] = rate.NewLimiter(rate.Limit(config.RateLimitCount), config.RateLimitDuration)
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
