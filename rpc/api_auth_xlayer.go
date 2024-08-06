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
)

// ApikeyAllowMap is the struct definition for the allowed API keys
type ApikeyAllowMap struct {
	Enable    bool
	AllowKeys map[string]ApiKeyItem
	sync.RWMutex
}

// ApiKeyItem is the struct containing the the API key data
type ApiKeyItem struct {
	Project string
	Timeout time.Time
}

// gApikeyAllowMap is the node's singleton instance for the allowed API keys
var gApikeyAllowMap = &ApikeyAllowMap{
	Enable:    false,
	AllowKeys: make(map[string]ApiKeyItem),
}

// InitApiAuth initializes the node API auth with the API key configs
func InitApiAuth(cfg string) {
	setApiAuth(cfg)
}

// setApiAuth sets the node API auth with the API key configs
func setApiAuth(cfg string) {
	gApikeyAllowMap.Lock()
	defer gApikeyAllowMap.Unlock()

	if cfg == "" {
		return
	}
	log.Info(fmt.Sprintf("API keys auth enabled, config: %v", cfg))
	keyItems := strings.Split(cfg, "\n")

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
			log.Warn(fmt.Sprintf("invalid key item: %s", item))
			continue
		}

		// Validate API key cfg inputs
		parse, err := time.Parse("2006-01-02", keyCfg.Timeout)
		if err != nil {
			log.Warn(fmt.Sprintf("failed to parse API key timeout cfg: %v, err: %v", keyCfg.Timeout, err))
			continue
		}
		if strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(keyCfg.Project+keyCfg.Timeout)))) != keyCfg.Key {
			log.Warn(fmt.Sprintf("project [%s], key [%s] is invalid, key = md5(Project+Timeout)", keyCfg.Project, keyCfg.Key))
			continue
		}
		// Set API key authentication
		key := strings.ToLower(keyCfg.Key)
		gApikeyAllowMap.AllowKeys[key] = ApiKeyItem{
			Project: keyCfg.Project,
			Timeout: parse,
		}
		// Set API key rate limiter
		rlCfg := RateLimitConfig{
			RateLimitApis:   keyCfg.Methods,
			RateLimitCount:  keyCfg.Count,
			RateLimitBucket: keyCfg.Bucket,
		}
		setApikeyRateLimit(key, rlCfg)
		gApikeyAllowMap.Enable = true
	}

}

// check returns the API key authentication check result
func check(key string) error {
	gApikeyAllowMap.RLock()
	defer gApikeyAllowMap.RUnlock()

	key = strings.ToLower(key)
	if item, ok := gApikeyAllowMap.AllowKeys[key]; ok && time.Now().Before(item.Timeout) {
		//metrics.RequestAuthCount(al.allowKeys[key].project)
		return nil
	} else if ok && time.Now().After(item.Timeout) {
		log.Warn(fmt.Sprintf("project [%s], key [%s] has expired, ", item.Project, key))
		//metrics.RequestAuthErrorCount(metrics.RequestAuthErrorTypeKeyExpired)
		return errors.New("key has expired")
	}
	//metrics.RequestAuthErrorCount(metrics.RequestAuthErrorTypeNoAuth)
	return errors.New("no authentication")
}

func apiAuthHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if gApikeyAllowMap.Enable {
			if er := check(path.Base(r.URL.Path)); er != nil {
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
