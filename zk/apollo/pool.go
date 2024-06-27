package apollo

import (
	"fmt"
	"os"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/log/v3"
)

func (c *Client) loadPool(value interface{}) {
	nodeCfg, ethCfg, err := c.unmarshal(value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to unmarshal config: %v", err))
		os.Exit(1)
	}

	// TODO: Switch to loading only pool configs
	c.ethCfg = ethCfg
	c.nodeCfg = nodeCfg
	log.Info(fmt.Sprintf("loaded pool from apollo config: %+v", value.(string)))
}

// firePool fires the pool config change
func (c *Client) firePool(key string, value *storage.ConfigChange) {
	nodeCfg, ethCfg, err := c.unmarshal(value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to unmarshal config: %v", err))
		return
	}

	log.Info(fmt.Sprintf("apollo eth backend old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo eth backend config changed: %+v", value.NewValue.(string)))

	log.Info(fmt.Sprintf("apollo node old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo node config changed: %+v", value.NewValue.(string)))

	nodecfg.UpdatePoolConfig(*nodeCfg)
	ethconfig.UpdatePoolConfig(*ethCfg)
}
