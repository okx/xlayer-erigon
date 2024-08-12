package apollo

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/erigon/rpc"
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

	// Set rpc flag on fire configuration changes
	setJsonRPCFlag()
}

// loadJsonRPCConfig loads the dynamic json rpc apollo configurations
func loadJsonRPCConfig(ctx *cli.Context) {
	unsafeGetApolloConfig().Lock()
	defer unsafeGetApolloConfig().Unlock()

	loadNodeJsonRPCConfig(ctx, &unsafeGetApolloConfig().NodeCfg)
	loadEthJsonRPCConfig(ctx, &unsafeGetApolloConfig().EthCfg)
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
	if ctx.IsSet(utils.RpcBatchLimit.Name) {
		nodeCfg.Http.BatchLimit = ctx.Int(utils.RpcBatchLimit.Name)
	}
	if ctx.IsSet(utils.WSEnabledFlag.Name) {
		nodeCfg.Http.WebsocketEnabled = true
	}
	if ctx.IsSet(utils.HTTPApiKeysFlag.Name) {
		nodeCfg.Http.HttpApiKeys = ctx.String(utils.HTTPApiKeysFlag.Name)
		rpc.SetApiAuth(nodeCfg.Http.HttpApiKeys)
	}
	if ctx.IsSet(utils.MethodRateLimitFlag.Name) {
		nodeCfg.Http.MethodRateLimit = ctx.String(utils.MethodRateLimitFlag.Name)
		rpc.SetRateLimit(nodeCfg.Http.MethodRateLimit)
	}
}

// loadEthJsonRPCConfig loads the dynamic json rpc apollo eth configurations
func loadEthJsonRPCConfig(ctx *cli.Context, ethCfg *ethconfig.Config) {
	// Load generic ZK config
	loadZkConfig(ctx, ethCfg)

	// Load jsonrpc config
}

// setJsonRPCFlag sets the dynamic json rpc apollo flag
func setJsonRPCFlag() {
	unsafeGetApolloConfig().Lock()
	defer unsafeGetApolloConfig().Unlock()
	unsafeGetApolloConfig().setRPCFlag()
}
