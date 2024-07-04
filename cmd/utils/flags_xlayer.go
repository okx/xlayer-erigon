package utils

import (
	"fmt"
	"time"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/urfave/cli/v2"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
)

var (
	// X Layer nacos
	NacosURLsFlag = cli.StringFlag{
		Name:  "zkevm.nacos-urls",
		Usage: "Nacos urls.",
		Value: "",
	}
	NacosNamespaceIdFlag = cli.StringFlag{
		Name:  "zkevm.nacos-namespace-id",
		Usage: "Nacos namespace Id.",
		Value: "",
	}
	NacosApplicationNameFlag = cli.StringFlag{
		Name:  "zkevm.nacos-application-name",
		Usage: "Nacos application name",
		Value: "",
	}
	NacosExternalListenAddrFlag = cli.StringFlag{
		Name:  "zkevm.nacos-external-listen-addr",
		Usage: "Nacos external listen addr.",
		Value: "",
	}
	TxPoolEnableWhitelistFlag = cli.BoolFlag{
		Name:  "txpool.enable.whitelist",
		Usage: "Enable or disable tx sender white list",
		Value: false,
	}
	TxPoolWhiteList = cli.StringFlag{
		Name:  "txpool.whitelist",
		Usage: "Comma separated list of addresses, who can send transactions",
		Value: "",
	}
	TxPoolBlockedList = cli.StringFlag{
		Name:  "txpool.blockedlist",
		Usage: "Comma separated list of addresses, who can't send and receive transactions",
		Value: "",
	}
)

func setGPOXLayer(ctx *cli.Context, cfg *gaspricecfg.XLayerConfig) {
	if ctx.IsSet(GpoTypeFlag.Name) {
		cfg.Type = ctx.String(GpoTypeFlag.Name)
	}

	if ctx.IsSet(GpoUpdatePeriodFlag.Name) {
		period, err := time.ParseDuration(ctx.String(GpoUpdatePeriodFlag.Name))
		if err != nil {
			panic(fmt.Sprintf("could not parse GpoUpdatePeriodFlag value %s", ctx.String(GpoUpdatePeriodFlag.Name)))
		}
		cfg.UpdatePeriod = period
	}

	if ctx.IsSet(GpoFactorFlag.Name) {
		cfg.Factor = ctx.Float64(GpoFactorFlag.Name)
	}

	if ctx.IsSet(GpoKafkaURLFlag.Name) {
		cfg.KafkaURL = ctx.String(GpoKafkaURLFlag.Name)
	}

	if ctx.IsSet(GpoTopicFlag.Name) {
		cfg.Topic = ctx.String(GpoTopicFlag.Name)
	}

	if ctx.IsSet(GpoGroupIDFlag.Name) {
		cfg.GroupID = ctx.String(GpoGroupIDFlag.Name)
	}

	if ctx.IsSet(GpoUsernameFlag.Name) {
		cfg.Username = ctx.String(GpoUsernameFlag.Name)
	}

	if ctx.IsSet(GpoPasswordFlag.Name) {
		cfg.Password = ctx.String(GpoPasswordFlag.Name)
	}

	if ctx.IsSet(GpoRootCAPathFlag.Name) {
		cfg.RootCAPath = ctx.String(GpoRootCAPathFlag.Name)
	}

	if ctx.IsSet(GpoL1CoinIdFlag.Name) {
		cfg.L1CoinId = ctx.Int(GpoL1CoinIdFlag.Name)
	}

	if ctx.IsSet(GpoL2CoinIdFlag.Name) {
		cfg.L2CoinId = ctx.Int(GpoL2CoinIdFlag.Name)
	}

	if ctx.IsSet(GpoDefaultL1CoinPriceFlag.Name) {
		cfg.DefaultL1CoinPrice = ctx.Float64(GpoDefaultL1CoinPriceFlag.Name)
	}

	if ctx.IsSet(GpoDefaultL2CoinPriceFlag.Name) {
		cfg.DefaultL2CoinPrice = ctx.Float64(GpoDefaultL2CoinPriceFlag.Name)
	}

	if ctx.IsSet(GpoGasPriceUsdtFlag.Name) {
		cfg.GasPriceUsdt = ctx.Float64(GpoGasPriceUsdtFlag.Name)
	}

	if ctx.IsSet(GpoEnableFollowerAdjustByL2L1PriceFlag.Name) {
		cfg.EnableFollowerAdjustByL2L1Price = ctx.Bool(GpoEnableFollowerAdjustByL2L1PriceFlag.Name)
	}

	if ctx.IsSet(GpoCongestionThresholdFlag.Name) {
		cfg.CongestionThreshold = ctx.Int(GpoCongestionThresholdFlag.Name)
	}
}

func setTxPoolXLayer(ctx *cli.Context, cfg *ethconfig.DeprecatedTxPoolConfig) {
	if ctx.IsSet(TxPoolEnableWhitelistFlag.Name) {
		cfg.EnableWhitelist = ctx.Bool(TxPoolEnableWhitelistFlag.Name)
	}
	if ctx.IsSet(TxPoolWhiteList.Name) {
		// Parse the command separated flag
		addrHexes := SplitAndTrim(ctx.String(TxPoolWhiteList.Name))
		cfg.WhiteList = make([]string, len(addrHexes))
		for i, senderHex := range addrHexes {
			sender := libcommon.HexToAddress(senderHex)
			cfg.WhiteList[i] = sender.String()
		}
	}
	if ctx.IsSet(TxPoolBlockedList.Name) {
		// Parse the command separated flag
		addrHexes := SplitAndTrim(ctx.String(TxPoolBlockedList.Name))
		cfg.BlockedList = make([]string, len(addrHexes))
		for i, senderHex := range addrHexes {
			sender := libcommon.HexToAddress(senderHex)
			cfg.BlockedList[i] = sender.String()
		}
	}
}
