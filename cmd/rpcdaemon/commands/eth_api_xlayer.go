package commands

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/common/math"
)

func (apii *APIImpl) GetGPCache() *GasPriceCache {
	return apii.gasCache
}

func (apii *APIImpl) runL2GasPricerForXLayer() {
	// set default gas price
	apii.gasCache.SetLatest(common.Hash{}, apii.L2GasPricer.GetConfig().Default)
	apii.gasCache.SetLatestRawGP(apii.L2GasPricer.GetConfig().Default)
	go apii.runL2GasPriceSuggester()
}

const cacheRGPDuration = 5 * time.Minute

type RawGPCache struct {
	mu     sync.Mutex
	values map[time.Time]*big.Int
}

func NewRawGPCache() *RawGPCache {
	return &RawGPCache{
		values: make(map[time.Time]*big.Int),
	}
}

// Add adds an RGP to the cache and removes old RGPs
func (c *RawGPCache) Add(rgp *big.Int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add new RGP with its timestamp
	timestamp := time.Now()
	c.values[timestamp] = new(big.Int).Set(rgp)

	// Remove values older than 5 minutes
	cleanupTime := time.Now().Add(-cacheRGPDuration)
	for t := range c.values {
		if t.Before(cleanupTime) {
			delete(c.values, t)
		}
	}
}

// GetMin returns the minimum RGP in the cache from the last few minutes
func (c *RawGPCache) GetMin() (*big.Int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.values) == 0 {
		return nil, fmt.Errorf("no values in cache")
	}

	minRGP := big.NewInt(0).SetInt64(math.MaxInt64) // Initialize to maximum big.Int
	for _, value := range c.values {
		if value.Cmp(minRGP) < 0 {
			minRGP = value
		}
	}
	return new(big.Int).Set(minRGP), nil
}

var GasPricerOnce sync.Once
