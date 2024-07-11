package commands

import (
	"github.com/gateway-fm/cdk-erigon-lib/common"
)

func (apii *APIImpl) GetGPCache() *GasPriceCache {
	return apii.gasCache
}

func (apii *APIImpl) runL2GasPricerForXLayer() {
	// set default gas price
	apii.gasCache.SetLatest(common.Hash{}, apii.L2GasPricer.GetConfig().Default)
	go apii.runL2GasPriceSuggester()
}
