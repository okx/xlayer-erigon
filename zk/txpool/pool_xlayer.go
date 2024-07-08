package txpool

import (
	"math/big"

	"github.com/gateway-fm/cdk-erigon-lib/common"
)

// WBConfig white and block config
type WBConfig struct {
	// BlockedList is the blocked address list
	BlockedList []string
	// EnableWhitelist is a flag to enable/disable the whitelist
	EnableWhitelist bool
	// WhiteList is the white address list
	WhiteList []string
}

type GPCache interface {
	GetLatest() (common.Hash, *big.Int)
	SetLatest(hash common.Hash, price *big.Int)
}

func (p *TxPool) checkBlockedAddr(addr common.Address) bool {
	// check from config
	for _, e := range p.wbCfg.BlockedList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) checkWhiteAddr(addr common.Address) bool {
	// check from config
	for _, e := range p.wbCfg.WhiteList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) SetGpCacheForXLayer(gpCache GPCache) {
	p.gpCache = gpCache
}
