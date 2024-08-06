package cli

import (
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/urfave/cli/v2"
)

func ApplyFlagsForEthXLayerConfig(ctx *cli.Context, cfg *ethconfig.Config) {
	cfg.XLayer = ethconfig.XLayerConfig{
		Apollo: ethconfig.ApolloClientConfig{
			Enable:        ctx.Bool(utils.ApolloEnableFlag.Name),
			IP:            ctx.String(utils.ApolloIPAddr.Name),
			AppID:         ctx.String(utils.ApolloAppId.Name),
			NamespaceName: ctx.String(utils.ApolloNamespaceName.Name),
		},
		Nacos: ethconfig.NacosConfig{
			URLs:               ctx.String(utils.NacosURLsFlag.Name),
			NamespaceId:        ctx.String(utils.NacosNamespaceIdFlag.Name),
			ApplicationName:    ctx.String(utils.NacosApplicationNameFlag.Name),
			ExternalListenAddr: ctx.String(utils.NacosExternalListenAddrFlag.Name),
		},
		EnableInnerTx:               ctx.Bool(utils.AllowInternalTransactions.Name),
		SequencerBatchSleepDuration: ctx.Duration(utils.SequencerBatchSleepDuration.Name),
	}
}

func ApplyFlagsForNodeXLayerConfig(ctx *cli.Context, cfg *nodecfg.Config) {
	cfg.Http.HttpApiKeys = ctx.String(utils.HTTPApiKeysFlag.Name)
	cfg.Http.MethodRateLimit = ctx.String(utils.MethodRateLimitFlag.Name)
}
