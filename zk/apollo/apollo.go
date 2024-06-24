package apollo

import (
	"fmt"
	"os"
	"strings"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/log/v3"
)

// Client is the apollo client
type Client struct {
	*agollo.Client
	config *ethconfig.ApolloConfig
}

// NewClient creates a new apollo client
func NewClient(conf *ethconfig.ApolloConfig) *Client {
	if conf == nil || !conf.Enable || conf.IP == "" || conf.AppID == "" || conf.NamespaceName == "" {
		log.Info(fmt.Sprintf("apollo is not enabled, config: %+v", conf))
		return nil
	}
	c := &config.AppConfig{
		IP:             conf.IP,
		AppID:          conf.AppID,
		NamespaceName:  conf.NamespaceName,
		Cluster:        "default",
		IsBackupConfig: false,
	}

	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		log.Error(fmt.Sprintf("failed init apollo: %v", err))
		os.Exit(1)
	}

	apc := &Client{
		Client: client,
		config: conf,
	}
	client.AddChangeListener(&CustomChangeListener{apc})

	return apc
}

// LoadConfig loads the config
func (c *Client) LoadConfig() (loaded bool) {
	if c == nil {
		return false
	}
	namespaces := strings.Split(c.config.NamespaceName, ",")
	for _, namespace := range namespaces {
		cache := c.GetConfigCache(namespace)
		cache.Range(func(key, value interface{}) bool {
			loaded = true
			switch namespace {
			case L2GasPricer:
				c.loadL2GasPricer(value)
			case JsonRPCRO, JsonRPCExplorer, JsonRPCSubgraph, JsonRPCLight, JsonRPCBridge, JsonRPCWO, JsonRPCUnlimited:
				c.loadJsonRPC(value)
			case Sequencer:
				c.loadSequencer(value)
			case Pool:
				c.loadPool(value)
			}
			return true
		})
	}
	return
}

// CustomChangeListener is the custom change listener
type CustomChangeListener struct {
	*Client
}

// OnChange is the change listener
func (c *CustomChangeListener) OnChange(changeEvent *storage.ChangeEvent) {
	for key, value := range changeEvent.Changes {
		if value.ChangeType == storage.MODIFIED {
			switch changeEvent.Namespace {
			case L2GasPricerHalt, SequencerHalt, JsonRPCROHalt, JsonRPCExplorerHalt, JsonRPCSubgraphHalt, JsonRPCLightHalt, JsonRPCBridgeHalt, JsonRPCWOHalt, JsonRPCUnlimitedHalt:
				c.fireHalt(key, value)
			case L2GasPricer:
				c.fireL2GasPricer(key, value)
			case Sequencer:
				c.fireSequencer(key, value)
			case JsonRPCRO, JsonRPCExplorer, JsonRPCSubgraph, JsonRPCLight, JsonRPCBridge, JsonRPCWO, JsonRPCUnlimited:
				c.fireJsonRPC(key, value)
			case Pool:
				c.firePool(key, value)
			}
		}
	}
}

// OnNewestChange is the newest change listener
func (c *CustomChangeListener) OnNewestChange(event *storage.FullChangeEvent) {
}
