package gasprice

import (
	"context"
	"math/big"
	"testing"
	"time"

	. "github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
)

func TestUpdateGasPriceFixed(t *testing.T) {
	ctx := context.Background()
	var d time.Duration = 1000000000
	cfg := Config{
		Default:  new(big.Int).SetUint64(1000000000),
		MaxPrice: new(big.Int).SetUint64(0),
		XLayer: XLayerConfig{
			Type:               FixedType,
			UpdatePeriod:       d,
			Factor:             0.5,
			KafkaURL:           "127.0.0.1:9092",
			Topic:              "middle_coinPrice_push",
			DefaultL2CoinPrice: 40,
			GasPriceUsdt:       0.001,
		},
	}
	l1GasPrice := big.NewInt(10000000000)
	f := newFixedGasPriceSuggester(ctx, cfg)

	f.UpdateGasPriceAvg(l1GasPrice)
}

func TestUpdateGasPriceAvgCases(t *testing.T) {
	var d time.Duration = 1000000000
	testcases := []struct {
		cfg        Config
		l1GasPrice *big.Int
		l2GasPrice uint64
	}{
		{
			cfg: Config{
				Default:  new(big.Int).SetUint64(1000000000),
				MaxPrice: new(big.Int).SetUint64(0),
				XLayer: XLayerConfig{
					Type:               FixedType,
					UpdatePeriod:       d,
					KafkaURL:           "127.0.0.1:9092",
					Topic:              "middle_coinPrice_push",
					DefaultL2CoinPrice: 40,
					GasPriceUsdt:       0.001,
				},
			},
			l1GasPrice: big.NewInt(10000000000),
			l2GasPrice: uint64(25000000000000),
		},
		{
			cfg: Config{
				Default:  new(big.Int).SetUint64(1000000000),
				MaxPrice: new(big.Int).SetUint64(0),
				XLayer: XLayerConfig{
					Type:               FixedType,
					UpdatePeriod:       d,
					KafkaURL:           "127.0.0.1:9092",
					Topic:              "middle_coinPrice_push",
					DefaultL2CoinPrice: 1e-19,
					GasPriceUsdt:       0.001,
				},
			},
			l1GasPrice: big.NewInt(10000000000),
			l2GasPrice: uint64(25000000000000),
		},
		{ // the gas price less than the min gas price
			cfg: Config{
				Default:  new(big.Int).SetUint64(26000000000000),
				MaxPrice: new(big.Int).SetUint64(0),
				XLayer: XLayerConfig{
					Type:               FixedType,
					UpdatePeriod:       d,
					KafkaURL:           "127.0.0.1:9092",
					Topic:              "middle_coinPrice_push",
					DefaultL2CoinPrice: 40,
					GasPriceUsdt:       0.001,
				},
			},
			l1GasPrice: big.NewInt(10000000000),
			l2GasPrice: uint64(26000000000000),
		},
		{ // the gas price bigger than the max gas price
			cfg: Config{
				Default:  new(big.Int).SetUint64(1000000000000),
				MaxPrice: new(big.Int).SetUint64(23000000000000),
				XLayer: XLayerConfig{
					Type:               FixedType,
					UpdatePeriod:       d,
					KafkaURL:           "127.0.0.1:9092",
					Topic:              "middle_coinPrice_push",
					DefaultL2CoinPrice: 40,
					GasPriceUsdt:       0.001,
				},
			},
			l1GasPrice: big.NewInt(10000000000),
			l2GasPrice: uint64(23000000000000),
		},
		{
			cfg: Config{
				Default:  new(big.Int).SetUint64(1000000000),
				MaxPrice: new(big.Int).SetUint64(0),
				XLayer: XLayerConfig{
					UpdatePeriod:       d,
					KafkaURL:           "127.0.0.1:9092",
					Topic:              "middle_coinPrice_push",
					DefaultL2CoinPrice: 30,
					GasPriceUsdt:       0.001,
				},
			},
			l1GasPrice: big.NewInt(10000000000),
			l2GasPrice: uint64(33300000000000),
		},
		{
			cfg: Config{
				Default:  new(big.Int).SetUint64(10),
				MaxPrice: new(big.Int).SetUint64(0),
				XLayer: XLayerConfig{
					Type:               FixedType,
					UpdatePeriod:       d,
					KafkaURL:           "127.0.0.1:9092",
					Topic:              "middle_coinPrice_push",
					DefaultL2CoinPrice: 30,
					GasPriceUsdt:       1e-15,
				},
			},
			l1GasPrice: big.NewInt(10000000000),
			l2GasPrice: uint64(33),
		},
	}

	for _, tc := range testcases {
		ctx := context.Background()
		f := newFixedGasPriceSuggester(ctx, tc.cfg)
		f.UpdateGasPriceAvg(tc.l1GasPrice)
	}
}
