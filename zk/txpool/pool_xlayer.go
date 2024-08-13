package txpool

import (
	"math/big"

	"github.com/gateway-fm/cdk-erigon-lib/common"
)

// XLayerConfig contains the X Layer configs for the txpool
type XLayerConfig struct {
	// BlockedList is the blocked address list
	BlockedList []string
	// EnableWhitelist is a flag to enable/disable the whitelist
	EnableWhitelist bool
	// WhiteList is the white address list
	WhiteList []string
	// FreeClaimGasAddrs is the address list for free claimTx
	FreeClaimGasAddrs []string
	// GasPriceMultiple is the factor claim tx gas price should mul
	GasPriceMultiple uint64
}

type GPCache interface {
	GetLatest() (common.Hash, *big.Int)
	SetLatest(hash common.Hash, price *big.Int)
}

// ApolloConfig is the interface for the singleton apollo config instance.
// This design is necessary to prevent circular dependencies on the txpool
// with the apollo package
type ApolloConfig interface {
	CheckBlockedAddr(localBlockedList []string, addr common.Address) bool
	GetEnableWhitelist(localEnableWhitelist bool) bool
	CheckWhitelistAddr(localWhitelist []string, addr common.Address) bool
	CheckFreeClaimAddr(localFreeClaimGasAddrs []string, addr common.Address) bool
}

// SetApolloConfig sets the apollo config with the node's apollo config
// singleton instance
func (p *TxPool) SetApolloConfig(cfg ApolloConfig) {
	p.apolloCfg = cfg
}

func (p *TxPool) isFreeClaimAddrXLayer(senderID uint64) bool {
	addr, ok := p.senders.senderID2Addr[senderID]
	if !ok {
		return false
	}
	return p.apolloCfg.CheckFreeClaimAddr(p.xlayerCfg.FreeClaimGasAddrs, addr)
}

func (p *TxPool) SetGpCacheForXLayer(gpCache GPCache) {
	p.gpCache = gpCache
}
