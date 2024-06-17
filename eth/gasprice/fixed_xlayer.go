package gasprice

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
	"github.com/ledgerwatch/log/v3"
)

// FixedGasPrice struct
type FixedGasPrice struct {
	cfg       gaspricecfg.Config
	ctx       context.Context
	lastRawGP *big.Int
	ratePrc   *KafkaProcessor
}

// newFixedGasPriceSuggester inits l2 fixed price suggester.
func newFixedGasPriceSuggester(ctx context.Context, cfg gaspricecfg.Config) *FixedGasPrice {
	gps := &FixedGasPrice{
		cfg:     cfg,
		ctx:     ctx,
		ratePrc: newKafkaProcessor(cfg, ctx),
	}
	return gps
}

// UpdateGasPriceAvg updates the gas price.
func (f *FixedGasPrice) UpdateGasPriceAvg(l1RpcUrl string) {
	//todo:apollo

	// Get L1 gasprice
	l1GasPrice, err := GetL1GasPrice(l1RpcUrl)
	if err != nil {
		log.Error("cannot get l1 gas price. Skipping update...")
		return
	}

	l2CoinPrice := f.ratePrc.GetL2CoinPrice()
	if l2CoinPrice < minCoinPrice {
		log.Warn("the L2 native coin price too small...")
		return
	}
	res := new(big.Float).Mul(big.NewFloat(0).SetFloat64(f.cfg.GasPriceUsdt/l2CoinPrice), big.NewFloat(0).SetFloat64(OKBWei))
	// Store l2 gasPrice calculated
	result := new(big.Int)
	res.Int(result)
	minGasPrice := new(big.Int).Set(f.cfg.Default)
	if minGasPrice.Cmp(result) == 1 { // minGasPrice > result
		log.Warn("setting DefaultGasPriceWei for L2")
		result = minGasPrice
	}
	maxGasPrice := new(big.Int).Set(f.cfg.MaxPrice)
	if maxGasPrice.Int64() > 0 && result.Cmp(maxGasPrice) == 1 { // result > maxGasPrice
		log.Warn("setting MaxGasPriceWei for L2")
		result = maxGasPrice
	}
	var truncateValue *big.Int
	log.Debug("Full L2 gas price value: ", result, ". Length: ", len(result.String()), ". L1 gas price value: ", l1GasPrice)

	numLength := len(result.String())
	if numLength > 3 { //nolint:gomnd
		aux := "%0" + strconv.Itoa(numLength-3) + "d" //nolint:gomnd
		var ok bool
		value := result.String()[:3] + fmt.Sprintf(aux, 0)
		truncateValue, ok = new(big.Int).SetString(value, 10)
		if !ok {
			log.Error("error converting: ", truncateValue)
		}
	} else {
		truncateValue = result
	}
	log.Debug("Storing truncated L2 gas price: ", truncateValue, ", L2 native coin price: ", l2CoinPrice)
	if truncateValue != nil {
		log.Info("Set l2 raw gas price: ", truncateValue.Uint64())
		f.lastRawGP = truncateValue
	} else {
		log.Error("nil value detected. Skipping...")
	}
}

func (f *FixedGasPrice) GetLastRawGP() *big.Int {
	return f.lastRawGP
}

func (f *FixedGasPrice) GetConfig() gaspricecfg.Config {
	return f.cfg
}

func (f *FixedGasPrice) GetCtx() context.Context {
	return f.ctx
}
