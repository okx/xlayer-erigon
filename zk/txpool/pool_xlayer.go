package txpool

import (
	"math/big"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
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
	// EnableFreeGasByNonce enable free gas
	EnableFreeGasByNonce bool
	// FreeGasExAddrs is the ex address which can be free gas for the transfer receiver
	FreeGasExAddrs []string
	// FreeGasCountPerAddr is the count limit of free gas tx per address
	FreeGasCountPerAddr uint64
	// FreeGasLimit is the max gas allowed use to do a free gas tx
	FreeGasLimit uint64
	// For special project
	// EnableFreeGasList enable the special project of XLayer for free gas
	EnableFreeGasList  bool
	FreeGasFromNameMap map[string]string                 // map[from]projectName
	FreeGasList        map[string]*ethconfig.FreeGasInfo // map[projectName]FreeGasInfo
}

type GPCache interface {
	GetLatest() (common.Hash, *big.Int)
	SetLatest(hash common.Hash, price *big.Int)
	GetLatestRawGP() *big.Int
	SetLatestRawGP(rgp *big.Int)
}

// ApolloConfig is the interface for the singleton apollo config instance.
// This design is necessary to prevent circular dependencies on the txpool
// with the apollo package
type ApolloConfig interface {
	CheckBlockedAddr(localBlockedList []string, addr common.Address) bool
	GetEnableWhitelist(localEnableWhitelist bool) bool
	CheckWhitelistAddr(localWhitelist []string, addr common.Address) bool
	CheckFreeClaimAddr(localFreeClaimGasAddrs []string, addr common.Address) bool
	CheckFreeGasExAddr(localFreeGasExAddrs []string, addr common.Address) bool
}

// SetApolloConfig sets the apollo config with the node's apollo config
// singleton instance
func (p *TxPool) SetApolloConfig(cfg ApolloConfig) {
	p.apolloCfg = cfg
}

func (p *TxPool) SetGpCacheForXLayer(gpCache GPCache) {
	p.gpCache = gpCache
}

func (p *TxPool) checkFreeGasExAddrXLayer(senderID uint64) bool {
	addr, ok := p.senders.senderID2Addr[senderID]
	if !ok {
		return false
	}
	return p.apolloCfg.CheckFreeGasExAddr(p.xlayerCfg.FreeGasExAddrs, addr)
}

func (p *TxPool) checkFreeGasAddrXLayer(senderID uint64) (bool, bool) {
	addr, ok := p.senders.senderID2Addr[senderID]
	if !ok {
		return false, false
	}
	// is claim tx
	if p.apolloCfg.CheckFreeClaimAddr(p.xlayerCfg.FreeClaimGasAddrs, addr) {
		return true, true
	}
	free := p.freeGasAddrs[addr.String()]
	return free, false
}

func (p *TxPool) isFreeGasXLayer(senderID uint64) bool {
	free, _ := p.checkFreeGasAddrXLayer(senderID)
	return free
}
