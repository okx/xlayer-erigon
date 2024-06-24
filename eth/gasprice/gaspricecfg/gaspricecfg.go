package gaspricecfg

import (
	"math/big"
	"time"

	"github.com/ledgerwatch/erigon/params"
)

var DefaultIgnorePrice = big.NewInt(2 * params.Wei)

var (
	DefaultMaxPrice = big.NewInt(500 * params.GWei)
)

type Config struct {
	Blocks           int
	Percentile       int
	MaxHeaderHistory int
	MaxBlockHistory  int
	Default          *big.Int `toml:",omitempty"`
	MaxPrice         *big.Int `toml:",omitempty"`
	IgnorePrice      *big.Int `toml:",omitempty"`

	// XLayer config
	Type         string        `toml:",omitempty"`
	UpdatePeriod time.Duration `toml:",omitempty"`
	Factor       float64       `toml:",omitempty"`
	KafkaURL     string        `toml:",omitempty"`
	Topic        string        `toml:",omitempty"`
	GroupID      string        `toml:",omitempty"`
	Username     string        `toml:",omitempty"`
	Password     string        `toml:",omitempty"`
	RootCAPath   string        `toml:",omitempty"`
	L1CoinId     int           `toml:",omitempty"`
	L2CoinId     int           `toml:",omitempty"`
	// DefaultL1CoinPrice is the L1 token's coin price
	DefaultL1CoinPrice float64 `toml:",omitempty"`
	// DefaultL2CoinPrice is the native token's coin price
	DefaultL2CoinPrice float64 `toml:",omitempty"`
	GasPriceUsdt       float64 `toml:",omitempty"`

	// EnableFollowerAdjustByL2L1Price is dynamic adjust the factor through the L1 and L2 coins price in follower strategy
	EnableFollowerAdjustByL2L1Price bool `toml:",omitempty"`
}

const (
	// DefaultType default gas price from config is set.
	DefaultType string = "default"

	// FollowerType calculate the gas price basing on the L1 gasPrice.
	FollowerType string = "follower"

	// FixedType the gas price from config that the unit is usdt, XLayer config
	FixedType string = "fixed"
)
