package txpool

import "github.com/gateway-fm/cdk-erigon-lib/common"

// WBConfig white and block config
type WBConfig struct {
	// BlockedList is the blocked address list
	BlockedList []string
	// EnableWhitelist is a flag to enable/disable the whitelist
	EnableWhitelist bool
	// WhiteList is the white address list
	WhiteList []string

	// EnableFreeGasByNonce enable free gas
	EnableFreeGasByNonce bool
	// FreeGasExAddress is the ex address which can be free gas for the transfer receiver
	FreeGasExAddress []string
	// FreeGasCountPerAddr is the count limit of free gas tx per address
	FreeGasCountPerAddr uint64
	// FreeGasLimit is the max gas allowed use to do a free gas tx
	FreeGasLimit uint64
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

func (p *TxPool) checkFreeGasExAddress(senderID uint64) bool {
	addr, ok := p.senders.senderID2Addr[senderID]
	if !ok {
		return false
	}
	for _, e := range p.wbCfg.FreeGasExAddress {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) checkFreeGas(senderID uint64) (bool, bool) {
	addr, ok := p.senders.senderID2Addr[senderID]
	if !ok {
		return false, false
	}

	// is claim tx
	//for _, e := range p.wbCfg.FreeClaimGasAddr {
	//	if common.HexToAddress(e) == addr {
	//		return true, true
	//	}
	//}
	free := p.freeGasAddress[addr.String()]
	return free, false
}
