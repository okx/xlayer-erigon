package apollo

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/apolloconfig/agollo/v4/storage"
	"gopkg.in/yaml.v2"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/erigon/turbo/node"
	"github.com/ledgerwatch/log/v3"
)

func (c *Client) unmarshal(value interface{}) (*nodecfg.Config, *ethconfig.Config, error) {
	config := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(value.(string)), config)
	if err != nil {
		log.Error(fmt.Sprintf("failed to load config: %v error: %v", value, err))
		return nil, nil, err
	}

	// sets global flags to value in apollo config
	ctx := createMockContext(c.flags)
	for key, value := range config {
		if !ctx.IsSet(key) {
			if reflect.ValueOf(value).Kind() == reflect.Slice {
				sliceInterface := value.([]interface{})
				s := make([]string, len(sliceInterface))
				for i, v := range sliceInterface {
					s[i] = fmt.Sprintf("%v", v)
				}
				err := ctx.Set(key, strings.Join(s, ","))
				if err != nil {
					return nil, nil, fmt.Errorf("failed setting %s flag with values=%s error=%s", key, s, err)
				}
			} else {
				err := ctx.Set(key, fmt.Sprintf("%v", value))
				if err != nil {
					return nil, nil, fmt.Errorf("failed setting %s flag with value=%v error=%s", key, value, err)
				}
			}
		}
	}

	nodeCfg := node.NewNodConfigUrfave(ctx)
	ethCfg := node.NewEthConfigUrfave(ctx, nodeCfg)

	return nodeCfg, ethCfg, nil
}

const (
	// Halt is the key for L2GasPricer halt
	Halt         = "Halt"
	maxHaltDelay = 20
)

func (c *Client) fireHalt(key string, value *storage.ConfigChange) {
	switch key {
	case Halt:
		if value.OldValue.(string) != value.NewValue.(string) {
			random, _ := rand.Int(rand.Reader, big.NewInt(maxHaltDelay))
			delay := time.Second * time.Duration(random.Int64())
			log.Info(fmt.Sprintf("halt changed from %s to %s delay halt %v", value.OldValue.(string), value.NewValue.(string), delay))
			time.Sleep(delay)
			os.Exit(1)
		}
	}
}

func (c *Client) loadConfig(value interface{}) {
	nodeCfg, ethCfg, err := c.unmarshal(value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to unmarshal config: %v", err))
		os.Exit(1)
	}

	c.ethCfg = ethCfg
	c.nodeCfg = nodeCfg
	log.Info(fmt.Sprintf("loaded config from apollo config: %+v", value.(string)))
}

func (c *Client) fireConfig(key string, value *storage.ConfigChange) {
	nodeCfg, ethCfg, err := c.unmarshal(value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to unmarshal config: %v", err))
		return
	}

	log.Info(fmt.Sprintf("apollo eth backend old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo eth backend config changed: %+v", value.NewValue.(string)))

	log.Info(fmt.Sprintf("apollo node old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo node config changed: %+v", value.NewValue.(string)))

	c.nodeCfg = nodeCfg
	c.ethCfg = ethCfg
}

func (c *Client) LoadTestConfig() (loaded bool) {
	if c == nil {
		return false
	}
	namespaces := strings.Split(c.ethCfg.Zk.XLayer.Apollo.NamespaceName, ",")
	for _, namespace := range namespaces {
		cache := c.GetConfigCache(namespace)
		cache.Range(func(key, value interface{}) bool {
			loaded = true
			switch namespace {
			case Test:
				c.loadConfig(value)
			}
			return true
		})
	}
	return
}
