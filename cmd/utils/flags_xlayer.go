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
	// XLayer apollo
	ApolloEnableFlag = cli.BoolFlag{
		Name:  "zkevm.apollo-enabled",
		Usage: "Apollo enable flag.",
	}
	ApolloIPAddr = cli.StringFlag{
		Name:  "zkevm.apollo-ip-addr",
		Usage: "Apollo IP address.",
	}
	ApolloAppId = cli.StringFlag{
		Name:  "zkevm.apollo-app-id",
		Usage: "Apollo App ID.",
	}
	ApolloNamespaceName = cli.StringFlag{
		Name:  "zkevm.apollo-namespace-name",
		Usage: "Apollo namespace name.",
	}
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
	// Pool
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
	// Gas Price
	GpoDefaultGasPriceFlag = cli.Int64Flag{
		Name:  "gpo.default-price",
		Usage: "Default gas price will be recommended by gpo",
		Value: ethconfig.Defaults.GPO.Default.Int64(),
	}
	GpoTypeFlag = cli.StringFlag{
		Name:  "gpo.type",
		Usage: "raw gas price strategy type: default, follower, fixed",
		Value: "default",
	}
	GpoUpdatePeriodFlag = cli.StringFlag{
		Name:  "gpo.update-period",
		Usage: "raw gas price update period",
		Value: "10s",
	}
	GpoFactorFlag = cli.Float64Flag{
		Name:  "gpo.factor",
		Usage: "raw gas price facotr",
		Value: 0.15,
	}
	GpoKafkaURLFlag = cli.StringFlag{
		Name:  "gpo.kafka-url",
		Usage: "raw gas price kafka url",
		Value: "",
	}
	GpoTopicFlag = cli.StringFlag{
		Name:  "gpo.topic",
		Usage: "raw gas price topic",
		Value: "",
	}
	GpoGroupIDFlag = cli.StringFlag{
		Name:  "gpo.group-id",
		Usage: "raw gas price group id",
		Value: "",
	}
	GpoUsernameFlag = cli.StringFlag{
		Name:  "gpo.username",
		Usage: "raw gas price username",
		Value: "",
	}
	GpoPasswordFlag = cli.StringFlag{
		Name:  "gpo.password",
		Usage: "raw gas price password",
		Value: "",
	}
	GpoRootCAPathFlag = cli.StringFlag{
		Name:  "gpo.root-ca-path",
		Usage: "raw gas price root ca path",
		Value: "",
	}
	GpoL1CoinIdFlag = cli.IntFlag{
		Name:  "gpo.l1-coin-id",
		Usage: "raw gas price l1 coin id",
		Value: 0,
	}
	GpoL2CoinIdFlag = cli.IntFlag{
		Name:  "gpo.l2-coin-id",
		Usage: "raw gas price l2 coin id",
		Value: 0,
	}
	GpoDefaultL1CoinPriceFlag = cli.Float64Flag{
		Name:  "gpo.default-l1-coin-price",
		Usage: "raw gas price default l1 coin price",
		Value: 0,
	}
	GpoDefaultL2CoinPriceFlag = cli.Float64Flag{
		Name:  "gpo.default-l2-coin-price",
		Usage: "raw gas price default l2 coin price",
		Value: 0,
	}
	GpoGasPriceUsdtFlag = cli.Float64Flag{
		Name:  "gpo.gas-price-usdt",
		Usage: "raw gas price usdt",
		Value: 0,
	}
	GpoEnableFollowerAdjustByL2L1PriceFlag = cli.BoolFlag{
		Name:  "gpo.enable-follower-adjust",
		Usage: "enable dynamic adjust the factor through the L1 and L2 coins price in follower strategy",
		Value: true,
	}
	GpoCongestionThresholdFlag = cli.IntFlag{
		Name:  "gpo.congestion-threshold",
		Usage: "Used to determine whether pending tx has reached the threshold for congestion",
		Value: 0,
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
