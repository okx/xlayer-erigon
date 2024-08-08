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
	loadL2GasPricerConfig(ctx)
	log.Info(fmt.Sprintf("loaded l2gaspricer from apollo config: %+v", value.(string)))
}

// fireL2GasPricer fires the l2gaspricer config change
func (c *Client) fireL2GasPricer(key string, value *storage.ConfigChange) {
	ctx, err := c.getConfigContext(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire l2gaspricer from apollo config failed, err: %v", err))
		return
	}

	loadL2GasPricerConfig(ctx)
	log.Info(fmt.Sprintf("apollo l2gaspricer old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo l2gaspricer config changed: %+v", value.NewValue.(string)))
}

// loadL2GasPricerConfig loads the dynamic gas pricer apollo configurations
func loadL2GasPricerConfig(ctx *cli.Context) {
	UnsafeGetApolloConfig().Lock()
	defer UnsafeGetApolloConfig().Unlock()

	loadNodeL2GasPricerConfig(ctx, &UnsafeGetApolloConfig().NodeCfg)
	loadEthL2GasPricerConfig(ctx, &UnsafeGetApolloConfig().EthCfg)
	UnsafeGetApolloConfig().setGPFlag()
}

// loadNodeL2GasPricerConfig loads the dynamic gas pricer apollo node configurations
func loadNodeL2GasPricerConfig(ctx *cli.Context, nodeCfg *nodecfg.Config) {
	// Load l2gaspricer config
}

// loadEthL2GasPricerConfig loads the dynamic gas pricer apollo eth configurations
func loadEthL2GasPricerConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	// Load generic ZK config
	loadZkConfig(ctx, ethCfg)

	// Load l2gaspricer config
}
