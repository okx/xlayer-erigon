package main

import (
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/zk/apollo"
	"github.com/ledgerwatch/erigon/zk/metrics"
	"github.com/ledgerwatch/log/v3"
	"github.com/urfave/cli/v2"
)

func initRunForXLayer(ctx *cli.Context, ethCfg *ethconfig.Config) {
	apolloClient := apollo.NewClient(ethCfg)
	if apolloClient.LoadConfig() {
		log.Info("Apollo config loaded")
	}

	// Init metrics
	if ctx.Bool(utils.MetricsEnabledFlag.Name) {
		metrics.Init()
	}
}
