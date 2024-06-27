package apollo

import (
	"fmt"
	"os"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/log/v3"
)

func (c *Client) loadJsonRPC(value interface{}) {
	nodeCfg, ethCfg, err := c.unmarshal(value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to unmarshal config: %v", err))
		os.Exit(1)
	}

	// TODO: Check and switch to loading only JSON-RPC configs
	c.ethCfg = ethCfg
	c.nodeCfg.Http = nodeCfg.Http
	log.Info(fmt.Sprintf("loaded json-rpc from apollo config: %+v", value.(string)))
}

// fireJsonRPC fires the json-rpc config change
func (c *Client) fireJsonRPC(key string, value *storage.ConfigChange) {
	nodeCfg, ethCfg, err := c.unmarshal(value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to unmarshal config: %v", err))
		os.Exit(1)
	}

	log.Info(fmt.Sprintf("apollo eth backend old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo eth backend config changed: %+v", value.NewValue.(string)))

	log.Info(fmt.Sprintf("apollo node old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo node config changed: %+v", value.NewValue.(string)))

	nodecfg.UpdateRPCConfig(*nodeCfg)
	ethconfig.UpdateRPCConfig(*ethCfg)
}
