package cli

import (
	"crypto/md5"
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
	Project string `mapstructure:"Project"`
	// Key defines the key
	Key string `mapstructure:"Key"`
	// Timeout defines the timeout
	Timeout string `mapstructure:"Timeout"`
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
	keyItems := strings.Split(apikeysconfig, ",")
	var keys []ApiKeyItem
	for _, item := range keyItems {
		parties := strings.Split(item, ":")
		if len(parties) != 3 {
			log.Warn("invalid key item: %s", item)
			continue
		}
		keys = append(keys, ApiKeyItem{Project: parties[0], Key: parties[1], Timeout: parties[2]})

	}
	setApiAuth(keys)
}

// setApiAuth sets the api authentication
func setApiAuth(kis []ApiKeyItem) {
	al.enable = len(kis) > 0
	var tmp = make(map[string]keyItem)
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
	}
	al.allowKeys = tmp
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

func apiAuthHandler(cfg string, next http.Handler) http.Handler {
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
