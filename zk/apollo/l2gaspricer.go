package apollo

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/log/v3"
	"github.com/urfave/cli/v2"
)

func (c *Client) loadL2GasPricer(value interface{}) {
	ctx, err := c.getConfigContext(value)
	if err != nil {
		utils.Fatalf("load l2gaspricer from apollo config failed, err: %v", err)
	}

	// Load l2gaspricer config changes
	loadNodeL2GasPricerConfig(ctx, c.nodeCfg)
	loadEthL2GasPricerConfig(ctx, c.ethCfg)
	log.Info(fmt.Sprintf("loaded l2gaspricer from apollo config: %+v", value.(string)))
}

// fireL2GasPricer fires the l2gaspricer config change
func (c *Client) fireL2GasPricer(key string, value *storage.ConfigChange) {
	ctx, err := c.getConfigContext(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire l2gaspricer from apollo config failed, err: %v", err))
		return
	}

	log.Info(fmt.Sprintf("apollo eth backend old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo eth backend config changed: %+v", value.NewValue.(string)))

	log.Info(fmt.Sprintf("apollo node old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo node config changed: %+v", value.NewValue.(string)))

	// Update l2gaspricer node config changes
	nodecfg.UnsafeGetApolloConfig().Lock()
	nodecfg.UnsafeGetApolloConfig().EnableApollo = true
	loadNodeL2GasPricerConfig(ctx, &nodecfg.UnsafeGetApolloConfig().Conf)
	nodecfg.UnsafeGetApolloConfig().Unlock()

	// Update l2gaspricer eth config changes
	ethconfig.UnsafeGetApolloConfig().Lock()
	ethconfig.UnsafeGetApolloConfig().EnableApollo = true
	loadEthL2GasPricerConfig(ctx, &ethconfig.UnsafeGetApolloConfig().Conf)
	ethconfig.UnsafeGetApolloConfig().Unlock()
}

func loadNodeL2GasPricerConfig(ctx *cli.Context, nodeCfg *nodecfg.Config) {
	// Load l2gaspricer config
}

func loadEthL2GasPricerConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	// Load l2gaspricer config
}
