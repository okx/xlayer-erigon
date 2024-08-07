package txpool

import (
	"math/big"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/gateway-fm/cdk-erigon-lib/common/fixedgas"
	"github.com/gateway-fm/cdk-erigon-lib/kv"
	"github.com/gateway-fm/cdk-erigon-lib/types"
	"github.com/ledgerwatch/log/v3"
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
	// okPayAccountList is the ok pay bundler accounts address
	OkPayAccountList []string
	// OkPayGasLimitPerBlock is the block max gas limit for ok pay tx
	OkPayGasLimitPerBlock uint64
}

type GPCache interface {
	GetLatest() (common.Hash, *big.Int)
	SetLatest(hash common.Hash, price *big.Int)
}

func (p *TxPool) checkBlockedAddr(addr common.Address) bool {
	// check from config
	for _, e := range p.xlayerCfg.BlockedList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) checkWhiteAddr(addr common.Address) bool {
	// check from config
	for _, e := range p.xlayerCfg.WhiteList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) isFreeClaimAddr(senderID uint64) bool {
	addr, ok := p.senders.senderID2Addr[senderID]
	if !ok {
		return false
	}
	for _, e := range p.xlayerCfg.FreeClaimGasAddrs {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) SetGpCacheForXLayer(gpCache GPCache) {
	p.gpCache = gpCache
}

func (p *TxPool) isOkPayAddr(addr common.Address) bool {
	for _, e := range p.xlayerCfg.OkPayAccountList {
		if common.HexToAddress(e) == addr {
			return true
		}
	}
	return false
}

func (p *TxPool) bestOkPay(n uint16, txs *types.TxsRlp, tx kv.Tx, isLondon, isShanghai bool, availableGas uint64, toSkip mapset.Set[[32]byte]) (uint64, int, []*metaTx, error) {
	var toRemove []*metaTx
	best := p.pending.best
	count := 0

	for i := 0; count < int(n) && i < len(best.ms); i++ {
		// if we wouldn't have enough gas for a standard transaction then quit out early
		if availableGas < fixedgas.TxGas {
			break
		}

		mt := best.ms[i]

		if toSkip.Contains(mt.Tx.IDHash) {
			continue
		}

		if !isLondon && mt.Tx.Type == 0x2 {
			// remove ldn txs when not in london
			toRemove = append(toRemove, mt)
			toSkip.Add(mt.Tx.IDHash)
			continue
		}

		if mt.Tx.Gas >= transactionGasLimit {
			// Skip transactions with very large gas limit, these shouldn't enter the pool at all
			log.Debug("found a transaction in the pending pool with too high gas for tx - clear the tx pool")
			continue
		}
		rlpTx, sender, isLocal, err := p.getRlpLocked(tx, mt.Tx.IDHash[:])
		if err != nil {
			return availableGas, count, toRemove, err
		}
		if len(rlpTx) == 0 {
			toRemove = append(toRemove, mt)
			continue
		}

		if !p.isOkPayAddr(sender) {
			continue
		}

		// make sure we have enough gas in the caller to add this transaction.
		// not an exact science using intrinsic gas but as close as we could hope for at
		// this stage
		intrinsicGas, _ := CalcIntrinsicGas(uint64(mt.Tx.DataLen), uint64(mt.Tx.DataNonZeroLen), nil, mt.Tx.Creation, true, true, isShanghai)
		if intrinsicGas > availableGas {
			// we might find another TX with a low enough intrinsic gas to include so carry on
			continue
		}

		if intrinsicGas <= availableGas { // check for potential underflow
			availableGas -= intrinsicGas
		}

		txs.Txs[count] = rlpTx
		copy(txs.Senders.At(count), sender.Bytes())
		txs.IsLocal[count] = isLocal
		toSkip.Add(mt.Tx.IDHash)
		count++
	}

	return availableGas, count, toRemove, nil
}
