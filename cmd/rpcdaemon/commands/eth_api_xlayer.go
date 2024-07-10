package commands

import (
	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
)

func (apii *APIImpl) GetGPCache() *GasPriceCache {
	return apii.gasCache
}

func (apii *APIImpl) runL2GasPricerForXLayer() {
	// set default gas price
	defaultPrice := apii.L2GasPricer.GetConfig().Default
	if defaultPrice == nil || defaultPrice.Int64() <= 0 {
		defaultPrice = gaspricecfg.DefaultMinimumBaseFee
	}
	apii.gasCache.SetLatest(common.Hash{}, defaultPrice)
	go apii.runL2GasPriceSuggester()
}
