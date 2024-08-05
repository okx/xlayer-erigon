package gasprice

import (
	"context"
	"math/big"

	"github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
)

// DefaultGasPricer gas price from config is set.
type DefaultGasPricer struct {
	cfg       gaspricecfg.Config
	ctx       context.Context
	lastRawGP *big.Int
}

// newDefaultGasPriceSuggester init default gas price suggester.
func newDefaultGasPriceSuggester(ctx context.Context, cfg gaspricecfg.Config) *DefaultGasPricer {
	gpe := &DefaultGasPricer{
		ctx: ctx,
		cfg: cfg,
	}
	return gpe
}

// UpdateGasPriceAvg not needed for default strategy.
func (d *DefaultGasPricer) UpdateGasPriceAvg(l1gp *big.Int) {
	d.lastRawGP = d.cfg.Default
}

func (d *DefaultGasPricer) UpdateConfig(c gaspricecfg.Config) {
	d.cfg = c
}

func (d *DefaultGasPricer) GetLastRawGP() *big.Int {
	return d.lastRawGP
}

func (d *DefaultGasPricer) GetConfig() gaspricecfg.Config {
	return d.cfg
}

func (d *DefaultGasPricer) GetCtx() context.Context {
	return d.ctx
}
