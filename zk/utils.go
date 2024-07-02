package zk

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/log/v3"
)

// prints progress every 10 seconds
// returns a channel to send progress to, and a function to stop the printer routine
func ProgressPrinter(message string, total uint64) (chan uint64, func()) {
	progress := make(chan uint64)
	ctDone := make(chan bool)

	go func() {
		defer close(progress)
		defer close(ctDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		var pc uint64
		var pct uint64

		for {
			select {
			case newPc := <-progress:
				pc += newPc
				if total > 0 {
					pct = (pc * 100) / total
				}
			case <-ticker.C:
				if pc > 0 {
					log.Info(fmt.Sprintf("%s: %d/%d (%d%%)", message, pc, total, pct))
				}
			case <-ctDone:
				return
			}
		}
	}()

	return progress, func() { ctDone <- true }
}

// prints progress every 10 seconds
// returns a channel to send progress to, and a function to stop the printer routine
func ProgressPrinterWithoutTotal(message string) (chan uint64, func()) {
	progress := make(chan uint64)
	ctDone := make(chan bool)

	go func() {
		defer close(progress)
		defer close(ctDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		var pc uint64

		for {
			select {
			case newPc := <-progress:
				pc = newPc
			case <-ticker.C:
				if pc > 0 {
					log.Info(fmt.Sprintf("%s: %d", message, pc))
				}
			case <-ctDone:
				return
			}
		}
	}()

	return progress, func() { ctDone <- true }
}

// prints progress every 10 seconds
// returns a channel to send progress to, and a function to stop the printer routine
func ProgressPrinterWithoutValues(message string, total uint64) (chan uint64, func()) {
	progress := make(chan uint64)
	ctDone := make(chan bool)

	go func() {
		defer close(progress)
		defer close(ctDone)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		var pc uint64
		var pct uint64

		for {
			select {
			case newPc := <-progress:
				pc = newPc
				if total > 0 {
					pct = (pc * 100) / total
				}
			case <-ticker.C:
				if pc > 0 {
					log.Info(fmt.Sprintf("%s: (%d%%)", message, pct))
				}
			case <-ctDone:
				return
			}
		}
	}()

	return progress, func() { ctDone <- true }
}

type GasPriceCache struct {
	latestPrice *big.Int
	latestHash  common.Hash
	mtx         sync.Mutex
}

func NewGasPriceCache() *GasPriceCache {
	return &GasPriceCache{
		latestPrice: big.NewInt(0),
		latestHash:  common.Hash{},
	}
}

func (c *GasPriceCache) GetLatest() (common.Hash, *big.Int) {
	var hash common.Hash
	var price *big.Int
	c.mtx.Lock()
	hash = c.latestHash
	price = c.latestPrice
	c.mtx.Unlock()
	return hash, price
}

func (c *GasPriceCache) SetLatest(hash common.Hash, price *big.Int) {
	c.mtx.Lock()
	c.latestPrice = price
	c.latestHash = hash
	c.mtx.Unlock()
}
