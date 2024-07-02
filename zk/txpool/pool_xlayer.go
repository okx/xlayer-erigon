package txpool

import (
	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/gateway-fm/cdk-erigon-lib/kv"
	"github.com/gateway-fm/cdk-erigon-lib/kv/kvcache"
	"github.com/gateway-fm/cdk-erigon-lib/txpool/txpoolcfg"
	"github.com/gateway-fm/cdk-erigon-lib/types"
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/zk"
	"math/big"
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

func XlayerNew(newTxs chan types.Announcements, coreDB kv.RoDB, cfg txpoolcfg.Config, ethCfg *ethconfig.Config, cache kvcache.Cache, chainID uint256.Int, shanghaiTime *big.Int, londonBlock *big.Int, gpCache *zk.GasPriceCache) (*TxPool, error) {
	pool, err := New(newTxs, coreDB, cfg, ethCfg, cache, chainID, shanghaiTime, londonBlock)
	if pool != nil {
		pool.gpCache = gpCache
	}
	return pool, err
}
