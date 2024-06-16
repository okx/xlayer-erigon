package txpool

import "github.com/gateway-fm/cdk-erigon-lib/common"

func (p *TxPool) checkBlockedAddr(addr common.Address) bool {
	// check from config
	for _, e := range p.cfg.BlockedList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) checkWhiteAddr(addr common.Address) bool {
	// check from config
	for _, e := range p.cfg.WhiteList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}
