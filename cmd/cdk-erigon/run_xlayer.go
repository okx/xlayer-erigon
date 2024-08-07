package main

import (
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/zk/apollo"
	"github.com/ledgerwatch/log/v3"
)

func initRunForXLayer(ethCfg *ethconfig.Config) {
	apolloClient := apollo.NewClient(ethCfg)
	if apolloClient.LoadConfig() {
		log.Info("apollo config loaded")
	}
}
