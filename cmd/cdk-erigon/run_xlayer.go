package main

import (
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/erigon/zk/apollo"
	"github.com/ledgerwatch/log/v3"
)

func initRunForXLayer(ethCfg *ethconfig.Config, nodeCfg *nodecfg.Config) {
	apolloClient := apollo.NewClient(ethCfg, nodeCfg)
	if apolloClient.LoadConfig() {
		log.Info("apollo config loaded")
	}
}
