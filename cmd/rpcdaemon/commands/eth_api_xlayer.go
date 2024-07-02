package commands

import (
	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/gateway-fm/cdk-erigon-lib/gointerfaces/txpool"
	"github.com/gateway-fm/cdk-erigon-lib/kv"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
	"github.com/ledgerwatch/erigon/zk"
)

// NewEthAPI returns APIImpl instance
func NewEthAPIXLayer(gpCache *zk.GasPriceCache, base *BaseAPI, db kv.RoDB, eth rpchelper.ApiBackend, txPool txpool.TxpoolClient, mining txpool.MiningClient, gascap uint64, returnDataLimit int, ethCfg *ethconfig.Config) *APIImpl {

	apii := NewEthAPI(base, db, eth, txPool, mining, gascap, returnDataLimit, ethCfg)
	apii.gasCache = gpCache

	// set default gas price
	apii.gasCache.SetLatest(common.Hash{}, apii.L2GasPircer.GetConfig().Default)
	go apii.runL2GasPriceSuggester()

	return apii
}
