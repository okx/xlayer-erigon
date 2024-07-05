package txpool

import (
	"math/big"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
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
	GetLatest() (libcommon.Hash, *big.Int)
	SetLatest(hash libcommon.Hash, price *big.Int)
}

func (p *TxPool) checkBlockedAddr(addr libcommon.Address) bool {
	// check from config
	for _, e := range p.wbCfg.BlockedList {
		if libcommon.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) checkWhiteAddr(addr libcommon.Address) bool {
	// check from config
	for _, e := range p.wbCfg.WhiteList {
		if libcommon.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) SetGpCacheForXLayer(gpCache GPCache) {
	p.gpCache = gpCache
}
