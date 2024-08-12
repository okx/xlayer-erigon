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

// loadPool loads the apollo pool config cache on startup
func (c *Client) loadPool(value interface{}) {
	ctx, err := c.getConfigContext(value)
	if err != nil {
		utils.Fatalf("load pool from apollo config failed, err: %v", err)
	}

	// Load pool config changes
	loadPoolConfig(ctx)
	log.Info(fmt.Sprintf("loaded pool from apollo config: %+v", value.(string)))
}

// firePool fires the pool config change
func (c *Client) firePool(key string, value *storage.ConfigChange) {
	ctx, err := c.getConfigContext(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire pool from apollo config failed, err: %v", err))
		return
	}

	loadPoolConfig(ctx)
	log.Info(fmt.Sprintf("apollo pool old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo pool config changed: %+v", value.NewValue.(string)))

	// Set pool flag on fire configuration changes
	setPoolFlag()
}

// loadPoolConfig loads the dynamic pool apollo configurations
func loadPoolConfig(ctx *cli.Context) {
	unsafeGetApolloConfig().Lock()
	defer unsafeGetApolloConfig().Unlock()

	loadNodePoolConfig(ctx, &unsafeGetApolloConfig().NodeCfg)
	loadEthPoolConfig(ctx, &unsafeGetApolloConfig().EthCfg)
}

// loadNodePoolConfig loads the dynamic pool apollo node configurations
func loadNodePoolConfig(ctx *cli.Context, nodeCfg *nodecfg.Config) {
	// Load pool config
}

// loadEthL2GasPricerConfig loads the dynamic gas pricer apollo eth configurations
func loadEthPoolConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	// Load generic ZK config
	loadZkConfig(ctx, ethCfg)

	// Load deprecated pool config
	utils.SetApolloPoolXLayer(ctx, &ethCfg.DeprecatedTxPool)
}

func setPoolFlag() {
	unsafeGetApolloConfig().Lock()
	defer unsafeGetApolloConfig().Unlock()
	unsafeGetApolloConfig().setPoolFlag()
}
