package gasprice

import (
	"context"
	"math/big"
	"testing"
	"time"

	. "github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
	"github.com/stretchr/testify/require"
)

func TestFollowerUpdateGasPrice(t *testing.T) {
	ctx := context.Background()
	var d time.Duration = 1000000000

	cfg := Config{
		Default:  new(big.Int).SetUint64(1000000000),
		MaxPrice: new(big.Int).SetUint64(0),
		XLayer: XLayerConfig{
			Type:               FollowerType,
			UpdatePeriod:       d,
			KafkaURL:           "127.0.0.1:9092",
			Topic:              "middle_coinPrice_push",
			Factor:             0.5,
			DefaultL1CoinPrice: 1,
			DefaultL2CoinPrice: 1,
		},
	}
	l1GasPrice := big.NewInt(10000000000)
	f := newFollowerGasPriceSuggester(ctx, cfg)

	f.UpdateGasPriceAvg(l1GasPrice)

	// Calculate correct GP
	// Correct GP: L1/L2 ratio = 1, l1GasPrice * 1 * Factor
	correctGpVal := big.NewFloat(0).Mul(big.NewFloat(0).SetFloat64(cfg.XLayer.Factor), big.NewFloat(0).SetInt(l1GasPrice))
	correctGp := new(big.Int)
	correctGpVal.Int(correctGp)
	require.Equal(t, correctGp.Uint64(), f.GetLastRawGP().Uint64())
}

func TestMaxGasPriceLimit(t *testing.T) {
	ctx := context.Background()
	var d time.Duration = 1000000000

	cfg := Config{
		Default:  new(big.Int).SetUint64(100000000),
		MaxPrice: new(big.Int).SetUint64(50000000),
		XLayer: XLayerConfig{
			Type:               FollowerType,
			UpdatePeriod:       d,
			KafkaURL:           "127.0.0.1:9092",
			Topic:              "middle_coinPrice_push",
			Factor:             0.5,
			DefaultL1CoinPrice: 1,
			DefaultL2CoinPrice: 1,
		},
	}
	l1GasPrice := big.NewInt(1000000000)
	f := newFollowerGasPriceSuggester(ctx, cfg)
	f.UpdateGasPriceAvg(l1GasPrice)
	require.Equal(t, f.GetLastRawGP().Uint64(), cfg.MaxPrice.Uint64())
}
