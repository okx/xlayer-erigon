package apollo

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	erigoncli "github.com/ledgerwatch/erigon/turbo/cli"
	"github.com/ledgerwatch/log/v3"
	"github.com/urfave/cli/v2"
)

func (c *Client) loadJsonRPC(value interface{}) {
	ctx, err := c.getConfigContext(value)
	if err != nil {
		utils.Fatalf("load jsonrpc from apollo config failed, err: %v", err)
	}

	// Load jsonrpc config changes
	loadJsonRPCConfig(ctx)
	log.Info(fmt.Sprintf("loaded jsonrpc from apollo config: %+v", value.(string)))
}

// fireJsonRPC fires the jsonrpc config change
func (c *Client) fireJsonRPC(key string, value *storage.ConfigChange) {
	ctx, err := c.getConfigContext(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire jsonrpc from apollo config failed, err: %v", err))
		return
	}

	loadJsonRPCConfig(ctx)
	log.Info(fmt.Sprintf("apollo jsonrpc old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo jsonrpc config changed: %+v", value.NewValue.(string)))
}

// loadJsonRPCConfig loads the dynamic json rpc apollo configurations
func loadJsonRPCConfig(ctx *cli.Context) {
	// Update jsonrpc node config changes
	nodecfg.UnsafeGetApolloConfig().Lock()
	nodecfg.UnsafeGetApolloConfig().EnableApollo = true
	loadNodeJsonRPCConfig(ctx, &nodecfg.UnsafeGetApolloConfig().Conf)
	nodecfg.UnsafeGetApolloConfig().Unlock()

	// Update jsonrpc eth config changes
	ethconfig.UnsafeGetApolloConfig().Lock()
	ethconfig.UnsafeGetApolloConfig().EnableApollo = true
	loadEthJsonRPCConfig(ctx, &ethconfig.UnsafeGetApolloConfig().Conf)
	ethconfig.UnsafeGetApolloConfig().Unlock()
}

// loadNodeJsonRPCConfig loads the dynamic json rpc apollo node configurations
func loadNodeJsonRPCConfig(ctx *cli.Context, nodeCfg *nodecfg.Config) {
	// Load jsonrpc config
	if ctx.IsSet(utils.HTTPEnabledFlag.Name) {
		nodeCfg.Http.Enabled = ctx.Bool(utils.HTTPEnabledFlag.Name)
	}
	if ctx.IsSet(utils.HTTPListenAddrFlag.Name) {
		nodeCfg.Http.HttpListenAddress = ctx.String(utils.HTTPListenAddrFlag.Name)
	}
	if ctx.IsSet(utils.HTTPPortFlag.Name) {
		nodeCfg.Http.HttpPort = ctx.Int(utils.HTTPPortFlag.Name)
	}
	if ctx.IsSet(erigoncli.HTTPReadTimeoutFlag.Name) {
		nodeCfg.Http.HTTPTimeouts.ReadTimeout = ctx.Duration(erigoncli.HTTPReadTimeoutFlag.Name)
	}
	if ctx.IsSet(erigoncli.HTTPWriteTimeoutFlag.Name) {
		nodeCfg.Http.HTTPTimeouts.WriteTimeout = ctx.Duration(erigoncli.HTTPWriteTimeoutFlag.Name)
	}
	if ctx.IsSet(utils.RpcBatchConcurrencyFlag.Name) {
		nodeCfg.Http.RpcBatchConcurrency = ctx.Uint(utils.RpcBatchConcurrencyFlag.Name)
	}
	if ctx.IsSet(utils.RpcBatchEnabled.Name) {
		nodeCfg.Http.BatchEnabled = ctx.Bool(utils.RpcBatchEnabled.Name)
	}
	if ctx.IsSet(utils.RpcBatchLimit.Name) {
		nodeCfg.Http.BatchLimit = ctx.Int(utils.RpcBatchLimit.Name)
	}
	if ctx.IsSet(utils.WSEnabledFlag.Name) {
		nodeCfg.Http.WebsocketEnabled = true
	}
}

// loadEthJsonRPCConfig loads the dynamic json rpc apollo eth configurations
func loadEthJsonRPCConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	// Load generic ZK config
	loadZkConfig(ctx, ethCfg)

	// Load jsonrpc config
}
