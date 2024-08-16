package commands

import (
	"fmt"
	"math/big"
	"sync"

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

// cacheSize = 300sec (TTL) / 10sec (UpdatePeriod) = 30
const cacheSize = 30 // Circular buffer size

// minGPWindowSize defines the number of recent raw GPs to consider
// when calculating the MinGP from the circular buffer.
// This value represents the window size within the buffer.
const minGPWindowSize = 27

type RawGPCache struct {
	values [cacheSize]*big.Int // Circular buffer
	head   int                 // Points to the current head of the buffer
}

// NewRawGPCache initializes a RawGPCache with a fixed size circular buffer
func NewRawGPCache() *RawGPCache {
	return &RawGPCache{
		head: 0,
	}
}

// Add adds an RGP to the circular buffer and manages the head position
func (c *RawGPCache) Add(rgp *big.Int) {
	// Add the new RGP to the circular buffer
	c.values[c.head] = new(big.Int).Set(rgp)
	c.head = (c.head + 1) % cacheSize
}

// GetMin returns the minimum RGP in the circular buffer
func (c *RawGPCache) GetMin() (*big.Int, error) {
	isEmpty := true
	minRGP := big.NewInt(0).SetInt64(math.MaxInt64) // Initialize to maximum big.Int
	for _, value := range c.values {
		if value == nil {
			continue
		}
		isEmpty = false
		if value.Cmp(minRGP) < 0 {
			minRGP = value
		}
	}

	if isEmpty {
		return nil, fmt.Errorf("no values in cache")
	}

	return new(big.Int).Set(minRGP), nil
}

// GetMinGPMoreRecent returns the minimum RGP in the circular buffer for the last minGPWindowSize elements
func (c *RawGPCache) GetMinGPMoreRecent() (*big.Int, error) {
	isEmpty := true
	minRGP := big.NewInt(0).SetInt64(math.MaxInt64) // Initialize to maximum big.Int

	// Calculate the start index for the last minGPWindowSize elements
	start := (c.head - minGPWindowSize + cacheSize) % cacheSize

	for i := 0; i < minGPWindowSize; i++ {
		index := (start + i) % cacheSize
		value := c.values[index]
		if value == nil {
			continue
		}
		isEmpty = false
		if value.Cmp(minRGP) < 0 {
			minRGP = value
		}
	}

	if isEmpty {
		return nil, fmt.Errorf("no values in cache")
	}

	return new(big.Int).Set(minRGP), nil
}

var GasPricerOnce sync.Once
