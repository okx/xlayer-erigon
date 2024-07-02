package gasprice

import (
	"context"
	"math/big"
	"testing"
	"time"

	. "github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
)

func TestUpdateGasPriceFollower(t *testing.T) {
	ctx := context.Background()
	var d time.Duration = 1000000000

	cfg := Config{
		Type:         FollowerType,
		Default:      new(big.Int).SetUint64(1000000000),
		MaxPrice:     new(big.Int).SetUint64(0),
		UpdatePeriod: d,
		Factor:       0.5,
	}
	l1GasPrice := big.NewInt(10000000000)
	f := newFollowerGasPriceSuggester(ctx, cfg)

	f.UpdateGasPriceAvg(l1GasPrice)
}

func TestLimitMasGasPrice(t *testing.T) {
	ctx := context.Background()
	var d time.Duration = 1000000000

	cfg := Config{
		Type:         FollowerType,
		Default:      new(big.Int).SetUint64(100000000),
		MaxPrice:     new(big.Int).SetUint64(50000000),
		UpdatePeriod: d,
		Factor:       0.5,
	}
	l1GasPrice := big.NewInt(1000000000)
	f := newFollowerGasPriceSuggester(ctx, cfg)
	f.UpdateGasPriceAvg(l1GasPrice)
}
