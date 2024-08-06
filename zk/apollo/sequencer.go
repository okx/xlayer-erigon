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

func (c *Client) loadSequencer(value interface{}) {
	ctx, err := c.getConfigContext(value)
	if err != nil {
		utils.Fatalf("load sequencer from apollo config failed, err: %v", err)
	}

	// Load sequencer config changes
	loadSequencerConfig(ctx)
	log.Info(fmt.Sprintf("loaded sequencer from apollo config: %+v", value.(string)))
}

// fireSequencer fires the sequencer config change
func (c *Client) fireSequencer(key string, value *storage.ConfigChange) {
	ctx, err := c.getConfigContext(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire sequencer from apollo config failed, err: %v", err))
		return
	}

	loadSequencerConfig(ctx)
	log.Info(fmt.Sprintf("apollo sequencer old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo sequencer config changed: %+v", value.NewValue.(string)))
}

// loadSequencerConfig loads the dynamic sequencer apollo configurations
func loadSequencerConfig(ctx *cli.Context) {
	// Update sequencer node config changes
	nodecfg.UnsafeGetApolloConfig().Lock()
	nodecfg.UnsafeGetApolloConfig().EnableApollo = true
	loadNodeSequencerConfig(ctx, &nodecfg.UnsafeGetApolloConfig().Conf)
	nodecfg.UnsafeGetApolloConfig().Unlock()

	// Update sequencer eth config changes
	ethconfig.UnsafeGetApolloConfig().Lock()
	ethconfig.UnsafeGetApolloConfig().EnableApollo = true
	loadEthSequencerConfig(ctx, &ethconfig.UnsafeGetApolloConfig().Conf)
	ethconfig.UnsafeGetApolloConfig().Unlock()
}

// loadNodeSequencerConfig loads the dynamic sequencer apollo node configurations
func loadNodeSequencerConfig(ctx *cli.Context, nodeCfg *nodecfg.Config) {
	// Load sequencer config
}

// loadEthSequencerConfig loads the dynamic sequencer apollo eth configurations
func loadEthSequencerConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	// Load generic ZK config
	loadZkConfig(ctx, ethCfg)

	// Load sequencer config
	if ctx.IsSet(utils.SequencerBlockSealTime.Name) {
		ethCfg.Zk.SequencerBlockSealTime = ctx.Duration(utils.SequencerBlockSealTime.Name)
	}
	if ctx.IsSet(utils.SequencerBatchSealTime.Name) {
		ethCfg.Zk.SequencerBatchSealTime = ctx.Duration(utils.SequencerBatchSealTime.Name)
	}
	if ctx.IsSet(utils.SequencerNonEmptyBatchSealTime.Name) {
		ethCfg.Zk.SequencerNonEmptyBatchSealTime = ctx.Duration(utils.SequencerNonEmptyBatchSealTime.Name)
	}
}
