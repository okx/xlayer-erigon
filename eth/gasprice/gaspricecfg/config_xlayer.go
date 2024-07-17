package gaspricecfg

import (
	"math/big"
	"time"

	"github.com/ledgerwatch/erigon/params"
)

// XLayerConfig is the X Layer gas price config
type XLayerConfig struct {
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

	CongestionThreshold int `toml:",omitempty"`
}

var (
	DefaultXLayerConfig = XLayerConfig{
		Type:                            DefaultType,
		UpdatePeriod:                    10 * time.Second,
		Factor:                          0.01,
		KafkaURL:                        "0.0.0.0",
		Topic:                           "xlayer",
		GroupID:                         "xlayer",
		DefaultL1CoinPrice:              2000,
		DefaultL2CoinPrice:              50,
		GasPriceUsdt:                    0.000000476190476,
		EnableFollowerAdjustByL2L1Price: true,
		CongestionThreshold:             0,
	}
	DefaultXLayerPrice = big.NewInt(1 * params.GWei)
)

const (
	// DefaultType default gas price from config is set.
	DefaultType string = "default"

	// FollowerType calculate the gas price basing on the L1 gasPrice.
	FollowerType string = "follower"

	// FixedType the gas price from config that the unit is usdt, XLayer config
	FixedType string = "fixed"
)
