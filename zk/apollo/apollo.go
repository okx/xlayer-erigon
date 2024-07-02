package apollo

import (
	"fmt"
	"strings"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/urfave/cli/v2"

	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	erigoncli "github.com/ledgerwatch/erigon/turbo/cli"
	"github.com/ledgerwatch/erigon/turbo/debug"
	"github.com/ledgerwatch/log/v3"
)

// Client is the apollo client
type Client struct {
	*agollo.Client
	namespaceMap map[string]string
	ethCfg       *ethconfig.Config
	nodeCfg      *nodecfg.Config
	flags        []cli.Flag
}

// NewClient creates a new apollo client
func NewClient(cfg *ethconfig.Config, nodeCfg *nodecfg.Config) *Client {
	if cfg == nil || !cfg.Zk.XLayer.Apollo.Enable || cfg.Zk.XLayer.Apollo.IP == "" || cfg.Zk.XLayer.Apollo.AppID == "" || cfg.Zk.XLayer.Apollo.NamespaceName == "" {
		log.Info(fmt.Sprintf("apollo is not enabled, config: %+v", cfg))
		return nil
	}
	c := &config.AppConfig{
		IP:             cfg.Zk.XLayer.Apollo.IP,
		AppID:          cfg.Zk.XLayer.Apollo.AppID,
		NamespaceName:  cfg.Zk.XLayer.Apollo.NamespaceName,
		Cluster:        "default",
		IsBackupConfig: false,
	}

	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		utils.Fatalf("failed init apollo: %v", err)
	}

	nsMap := make(map[string]string)
	namespaces := strings.Split(cfg.Zk.XLayer.Apollo.NamespaceName, ",")
	for _, namespace := range namespaces {
		prefix, err := getNamespacePrefix(namespace)
		if err != nil {
			utils.Fatalf("failed init apollo: %v", err)
		}
		nsMap[prefix] = namespace
	}

	apc := &Client{
		Client:       client,
		namespaceMap: nsMap,
		ethCfg:       cfg,
		nodeCfg:      nodeCfg,
		flags:        append(erigoncli.DefaultFlags, debug.Flags...),
	}
	client.AddChangeListener(&CustomChangeListener{apc})

	return apc
}

// LoadConfig loads the config
func (c *Client) LoadConfig() (loaded bool) {
	if c == nil {
		return false
	}
	for prefix, namespace := range c.namespaceMap {
		cache := c.GetConfigCache(namespace)
		cache.Range(func(key, value interface{}) bool {
			loaded = true
			switch prefix {
			case Sequencer:
				c.loadSequencer(value)
			case JsonRPC:
				c.loadJsonRPC(value)
			case L2GasPricer:
				c.loadL2GasPricer(value)
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
			suffix, err := getNamespaceSuffix(changeEvent.Namespace)
			if err != nil {
				log.Warn(fmt.Sprintf("not processing change event: %v", err))
				continue
			}
			switch suffix {
			case Halt:
				c.fireHalt(key, value)
				continue
			}

			prefix, err := getNamespacePrefix(changeEvent.Namespace)
			if err != nil {
				log.Warn(fmt.Sprintf("not processing change event: %v", err))
				continue
			}
			switch prefix {
			case Sequencer:
				c.fireSequencer(key, value)
			case JsonRPC:
				c.fireJsonRPC(key, value)
			case L2GasPricer:
				c.fireL2GasPricer(key, value)
			}
		}
	}
}

// OnNewestChange is the newest change listener
func (c *CustomChangeListener) OnNewestChange(event *storage.FullChangeEvent) {
}
