package apollo

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/log/v3"
)

func (c *Client) loadJsonRPC(value interface{}) {
	nodeCfg, _, err := c.unmarshal(value)
	if err != nil {
		utils.Fatalf("load jsonrpc from apollo config failed, unmarshal err: %v", err)
	}

	// TODO: Add specific RPC configs to load from apollo config
	c.nodeCfg.Http = nodeCfg.Http
	log.Info(fmt.Sprintf("loaded jsonrpc from apollo config: %+v", value.(string)))
}

// fireJsonRPC fires the json-rpc config change
func (c *Client) fireJsonRPC(key string, value *storage.ConfigChange) {
	nodeCfg, ethCfg, err := c.unmarshal(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire jsonrpc from apollo config failed, unmarshal err: %v", err))
		return
	}

	log.Info(fmt.Sprintf("apollo eth backend old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo eth backend config changed: %+v", value.NewValue.(string)))

	log.Info(fmt.Sprintf("apollo node old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo node config changed: %+v", value.NewValue.(string)))

	nodecfg.UpdateRPCConfig(*nodeCfg)
	ethconfig.UpdateRPCConfig(*ethCfg)
}
