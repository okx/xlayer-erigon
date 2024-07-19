package rpc

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/types"
	"github.com/ledgerwatch/log/v3"
)

// ApiKeyItem is the api key item
type ApiKeyItem struct {
	// Name defines the name of the key
	Project string `json:"project"`
	// Key defines the key
	Key string `json:"key"`
	// Timeout defines the timeout
	Timeout string `json:"timeout"`
	// Methods defines the methods
	rateLimitConfig *RateLimitConfig
}

type apiAllow struct {
	allowKeys map[string]keyItem
	enable    bool
}

type keyItem struct {
	project string
	timeout time.Time
}

var al apiAllow

// InitApiAuth initializes the api authentication
func InitApiAuth(apikeysconfig string) {
	if apikeysconfig == "" {
		return
	}
	log.Info("api auth enabled", "apikeysconfig", apikeysconfig)
	keyItems := strings.Split(apikeysconfig, "\n")
	var keys []ApiKeyItem

	for _, item := range keyItems {
		var itemins = struct {
			// Name defines the name of the key
			Project string   `json:"project"`
			Key     string   `json:"key"`
			Timeout string   `json:"timeout"`
			Methods []string `json:"methods"`
			Count   int      `json:"count"`
			Bucket  int      `json:"bucket"`
		}{}
		err := json.Unmarshal([]byte(item), &itemins)
		if err != nil {
			log.Warn("invalid key item: %s", item)
			continue
		}
		apiKeyItem := ApiKeyItem{Project: itemins.Project, Key: itemins.Key, Timeout: itemins.Timeout}
		if len(itemins.Methods) > 0 {
			rlc := RateLimitConfig{
				RateLimitApis:   itemins.Methods,
				RateLimitCount:  itemins.Count,
				RateLimitBucket: itemins.Bucket,
			}
			apiKeyItem.rateLimitConfig = &rlc
		}
		keys = append(keys, apiKeyItem)
	}
	setApiAuth(keys)
}

// setApiAuth sets the api authentication
func setApiAuth(kis []ApiKeyItem) {
	al.enable = len(kis) > 0
	var tmp = make(map[string]keyItem)
	var rateLimitConfig = make(map[string]*RateLimitConfig)
	for _, k := range kis {
		k.Key = strings.ToLower(k.Key)
		parse, err := time.Parse("2006-01-02", k.Timeout)
		if err != nil {
			log.Warn("parse key [%+v], error parsing timeout: %v", k, err)
			continue
		}
		if strings.ToLower(fmt.Sprintf("%x", md5.Sum([]byte(k.Project+k.Timeout)))) != k.Key {
			log.Warn("project [%s], key [%s] is invalid, key = md5(Project+Timeout)", k.Project, k.Key)
			continue
		}
		tmp[k.Key] = keyItem{project: k.Project, timeout: parse}
		if k.rateLimitConfig != nil {
			rateLimitConfig[k.Key] = k.rateLimitConfig
		}
	}
	al.allowKeys = tmp
	initApikeyRateLimit(rateLimitConfig)
}

func check(key string) error {
	key = strings.ToLower(key)
	if item, ok := al.allowKeys[key]; ok && time.Now().Before(item.timeout) {
		//metrics.RequestAuthCount(al.allowKeys[key].project)
		return nil
	} else if ok && time.Now().After(item.timeout) {
		log.Warn("project [%s], key [%s] has expired, ", item.project, key)
		//metrics.RequestAuthErrorCount(metrics.RequestAuthErrorTypeKeyExpired)
		return errors.New("key has expired")
	}
	//metrics.RequestAuthErrorCount(metrics.RequestAuthErrorTypeNoAuth)
	return errors.New("no authentication")
}

func apiAuthHandlerFunc(cfg string, handlerFunc http.HandlerFunc) http.HandlerFunc {
	InitApiAuth(cfg)
	return func(w http.ResponseWriter, r *http.Request) {
		if al.enable {
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

func ApiAuthHandler(cfg string, next http.Handler) http.Handler {
	return apiAuthHandlerFunc(cfg, next.ServeHTTP)
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
