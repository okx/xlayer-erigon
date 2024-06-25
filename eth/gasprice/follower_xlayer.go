package gasprice

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
	"github.com/ledgerwatch/log/v3"
)

const (
	// OKBWei OKB wei
	OKBWei       = 1e18
	minCoinPrice = 1e-18
)

// FollowerGasPrice struct.
type FollowerGasPrice struct {
	cfg       gaspricecfg.Config
	ctx       context.Context
	lastRawGP *big.Int
	kafkaPrc  *KafkaProcessor
}

// newFollowerGasPriceSuggester inits l2 follower gas price suggester which is based on the l1 gas price.
func newFollowerGasPriceSuggester(ctx context.Context, cfg gaspricecfg.Config) *FollowerGasPrice {
	gps := &FollowerGasPrice{
		cfg:       cfg,
		ctx:       ctx,
		lastRawGP: new(big.Int).SetUint64(1),
	}
	if cfg.EnableFollowerAdjustByL2L1Price {
		gps.kafkaPrc = newKafkaProcessor(cfg, ctx)
	}

	return gps
}

// UpdateGasPriceAvg updates the gas price.
func (f *FollowerGasPrice) UpdateGasPriceAvg(l1GasPrice *big.Int) {
	//todo: apollo

	if big.NewInt(0).Cmp(l1GasPrice) == 0 {
		log.Warn("gas price 0 received. Skipping update...")
		return
	}

	// Apply factor to calculate l2 gasPrice
	factor := big.NewFloat(0).SetFloat64(f.cfg.Factor)
	res := new(big.Float).Mul(factor, big.NewFloat(0).SetInt(l1GasPrice))

	// convert the eth gas price to okb gas price
	if f.cfg.EnableFollowerAdjustByL2L1Price {
		l1CoinPrice, l2CoinPrice := f.kafkaPrc.GetL1L2CoinPrice()
		if l1CoinPrice < minCoinPrice || l2CoinPrice < minCoinPrice {
			log.Warn("the L1 or L2 native coin price too small...")
			return
		}
		res = new(big.Float).Mul(big.NewFloat(0).SetFloat64(l1CoinPrice/l2CoinPrice), res)
		log.Debug(fmt.Sprintf("L2 pre gas price value: %s. L1 coin price: %f. L2 coin price: %f", res.String(), l1CoinPrice, l2CoinPrice))
	}

	// Cache l2 gasPrice calculated
	result := new(big.Int)
	res.Int(result)
	minGasPrice := new(big.Int).Set(f.cfg.Default)
	if minGasPrice.Cmp(result) == 1 { // minGasPrice > result
		log.Warn(fmt.Sprintf("setting DefaultGasPrice for L2: %s", f.cfg.Default.String()))
		result = minGasPrice
	}
	maxGasPrice := new(big.Int).Set(f.cfg.MaxPrice)
	if maxGasPrice.Int64() > 0 && result.Cmp(maxGasPrice) == 1 { // result > maxGasPrice
		log.Warn("setting MaxGasPriceWei for L2")
		result = maxGasPrice
	}
	var truncateValue *big.Int
	log.Debug(fmt.Sprintf("Full L2 gas price value: %s. Length: %d", result.String(), len(result.String())))
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

	if truncateValue != nil {
		log.Info(fmt.Sprintf("Set l2 raw gas price: %d", truncateValue.Uint64()))
		f.lastRawGP = truncateValue
	} else {
		log.Error("nil value detected. Skipping...")
	}
}

func (f *FollowerGasPrice) GetLastRawGP() *big.Int {
	return f.lastRawGP
}

func (f *FollowerGasPrice) GetConfig() gaspricecfg.Config {
	return f.cfg
}

func (f *FollowerGasPrice) GetCtx() context.Context {
	return f.ctx
}
