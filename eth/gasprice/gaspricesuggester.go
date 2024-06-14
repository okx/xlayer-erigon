package gasprice

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ledgerwatch/erigon/eth/gasprice/gaspricecfg"
	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
	"github.com/ledgerwatch/log/v3"
)

// L2GasPricer interface for gas price suggester.
type L2GasPricer interface {
	UpdateGasPriceAvg(string)
	GetLastRawGP() *big.Int
	GetConfig() gaspricecfg.Config
	GetCtx() context.Context
}

// NewL2GasPriceSuggester init.
func NewL2GasPriceSuggester(ctx context.Context, cfg gaspricecfg.Config) L2GasPricer {
	var gpricer L2GasPricer
	switch cfg.Type {
	case gaspricecfg.FollowerType:
		log.Info("Follower type selected")
		gpricer = newFollowerGasPriceSuggester(ctx, cfg)
	//case DefaultType:
	//	log.Info("Default type selected")
	//	gpricer = newDefaultGasPriceSuggester(ctx, cfg, pool)
	//case FixedType:
	//	log.Info("Fixed type selected")
	//	gpricer = newFixedGasPriceSuggester(ctx, cfg, pool, ethMan)
	default:
		log.Error("unknown l2 gas price suggester type ", cfg.Type, ". Please specify a valid one: 'follower' or 'default'")
	}

	return gpricer
}

func GetL1GasPrice(l1RpcUrl string) (*big.Int, error) {
	res, err := client.JSONRPCCall(l1RpcUrl, "eth_gasPrice")
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, fmt.Errorf("RPC error response: %s", res.Error.Message)
	}
	if res.Error != nil {
		return nil, fmt.Errorf("RPC error response: %s", res.Error.Message)
	}

	var resultString string
	if err := json.Unmarshal(res.Result, &resultString); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %v", err)
	}

	price, ok := big.NewInt(0).SetString(resultString[2:], 16)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to big.Int")
	}

	return price, nil
}
