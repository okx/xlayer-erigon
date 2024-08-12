package rpc

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/types"
	"github.com/ledgerwatch/log/v3"
	"golang.org/x/time/rate"
)

// ApiKeyAutRateLimiter is the struct definition for the allowed API auth keys rate limiter
type ApiKeyAutRateLimiter struct {
	Enable    bool
	AllowKeys map[string]ApiKeyItem
	rlm       map[string]map[string]*rate.Limiter
	sync.RWMutex
}

// ApiKeyItem is the struct containing the the API key data
type ApiKeyItem struct {
	Project string
	Timeout time.Time
}

// gApiKeyAutRateLimiter is the node's singleton instance for the allowed API auth keys rate limiter
var gApiKeyAutRateLimiter = &ApiKeyAutRateLimiter{
	Enable:    false,
	AllowKeys: make(map[string]ApiKeyItem),
	rlm:       make(map[string]map[string]*rate.Limiter),
}

// SetApiAuth sets the gApiKeyAutRateLimiter singleton instance with the API
// auth key configs
func SetApiAuth(cfg string) {
	gApiKeyAutRateLimiter.Lock()
	defer gApiKeyAutRateLimiter.Unlock()

	if cfg == "" {
		return
	}
	log.Info(fmt.Sprintf("Setting API keys auth, config: %v", cfg))
	keyItems := strings.Split(cfg, "\n")

	// Clear API auth key maps
	gApiKeyAutRateLimiter.AllowKeys = make(map[string]ApiKeyItem)
	gApiKeyAutRateLimiter.rlm = make(map[string]map[string]*rate.Limiter)

	// Set API auth key map
	for _, item := range keyItems {
		var keyCfg = struct {
			// Name defines the name of the key
			Project string   `json:"project"`
			Key     string   `json:"key"`
			Timeout string   `json:"timeout"`
			Methods []string `json:"methods"`
			Count   int      `json:"count"`
			Bucket  int      `json:"bucket"`
		}{}
		err := json.Unmarshal([]byte(item), &keyCfg)
		if err != nil {
			log.Warn(fmt.Sprintf("Invalid key item: %s", item))
			continue
		}

		// Validate API key cfg inputs
		parse, err := time.Parse("2006-01-02", keyCfg.Timeout)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to parse API key timeout cfg: %v, err: %v", keyCfg.Timeout, err))
			continue
		}
		if strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(keyCfg.Project+keyCfg.Timeout)))) != keyCfg.Key {
			log.Warn(fmt.Sprintf("Project [%s], key [%s] is invalid, key = md5(Project+Timeout)", keyCfg.Project, keyCfg.Key))
			continue
		}

		// Set API key authentication
		key := strings.ToLower(keyCfg.Key)
		gApiKeyAutRateLimiter.AllowKeys[key] = ApiKeyItem{
			Project: keyCfg.Project,
			Timeout: parse,
		}

		// Set API key rate limiter
		if _, ok := gApiKeyAutRateLimiter.rlm[key]; ok {
			log.Warn("API key rate limiter already set, skipping")
			return
		}

		// Set API key rate limiter map
		gApiKeyAutRateLimiter.rlm[key] = make(map[string]*rate.Limiter)
		for _, api := range keyCfg.Methods {
			gApiKeyAutRateLimiter.rlm[key][api] = rate.NewLimiter(rate.Limit(keyCfg.Count), keyCfg.Bucket)
			log.Info(fmt.Sprintf("Rate limiter enabled for key: %v for api method: %v with count: %v and bucket: %v", key, keyCfg.Methods, keyCfg.Count, keyCfg.Bucket))
		}
		log.Info(fmt.Sprintf("Set API key rate limiter for key: %v, cfg: %v", key, cfg))
		gApiKeyAutRateLimiter.Enable = true
	}
}

// checkApikeyMethodRateLimit returns true if the key and the method API is allowed
// by the API key rate limiter
func checkApikeyMethodRateLimit(key, method string) bool {
	gApiKeyAutRateLimiter.RLock()
	defer gApiKeyAutRateLimiter.RUnlock()

	if rlm, keyFound := gApiKeyAutRateLimiter.rlm[key]; keyFound {
		if rl, ok := rlm[method]; ok {
			return rl.Allow()
		}
	}
	return true
}

// checkAuthKey checks the API authentication key
func checkAuthKey(key string) error {
	gApiKeyAutRateLimiter.RLock()
	defer gApiKeyAutRateLimiter.RUnlock()

	key = strings.ToLower(key)
	if item, ok := gApiKeyAutRateLimiter.AllowKeys[key]; ok && time.Now().Before(item.Timeout) {
		//metrics.RequestAuthCount(al.allowKeys[key].project)
		return nil
	} else if ok && time.Now().After(item.Timeout) {
		log.Warn(fmt.Sprintf("Project [%s], key [%s] has expired, ", item.Project, key))
		//metrics.RequestAuthErrorCount(metrics.RequestAuthErrorTypeKeyExpired)
		return errors.New("key has expired")
	}
	//metrics.RequestAuthErrorCount(metrics.RequestAuthErrorTypeNoAuth)
	return errors.New("no authentication")
}

func apiAuthHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if gApiKeyAutRateLimiter.Enable {
			if er := checkAuthKey(path.Base(r.URL.Path)); er != nil {
				err := handleNoAuthErr(w, er)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}
		handlerFunc(w, r)
	}
}

func ApiAuthHandler(next http.Handler) http.Handler {
	return apiAuthHandlerFunc(next.ServeHTTP)
}

func handleNoAuthErr(w http.ResponseWriter, err error) error {
	respbytes, err := types.NewResponse(types.Request{JSONRPC: "2.0", ID: 0}, nil, types.NewRPCError(types.InvalidParamsErrorCode, err.Error())).Bytes()
	if err != nil {
		return err
	}
	_, err = w.Write(respbytes)
	if err != nil {
		return err
	}
	return nil
}
