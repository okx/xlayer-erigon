package commands

import (
	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/zk"
)

// AddGasCachePointer set gas price cache
func (apii *APIImpl) AddGasCachePointer(gpCache *zk.GasPriceCache) {
	apii.gasCache = gpCache

	// set default gas price
	apii.gasCache.SetLatest(common.Hash{}, apii.L2GasPircer.GetConfig().Default)
	go apii.runL2GasPriceSuggester()
}
