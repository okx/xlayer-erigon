package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ledgerwatch/erigon/common/hexutil"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/eth/gasprice"
	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
	"github.com/ledgerwatch/log/v3"
)

type L1GasPrice struct {
	timestamp time.Time
	gasPrice  *big.Int
}

func (api *APIImpl) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	cc, err := api.chainConfig(tx)
	if err != nil {
		return nil, err
	}
	chainId := cc.ChainID
	if !api.isZkNonSequencer(chainId) {
		return api.GasPrice_nonRedirected(ctx)
	}

	price, err := api.getGPFromTrustedNode()
	if err != nil {
		log.Error("eth_gasPrice error: ", err)
		return (*hexutil.Big)(api.L2GasPircer.GetConfig().Default), nil
	}

	return (*hexutil.Big)(price), nil
}

func (api *APIImpl) GasPrice_nonRedirected(ctx context.Context) (*hexutil.Big, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	cc, err := api.chainConfig(tx)
	if err != nil {
		return nil, err
	}
	oracle := gasprice.NewOracle(NewGasPriceOracleBackend(tx, cc, api.BaseAPI), ethconfig.Defaults.GPO, api.gasCache)
	tipcap, err := oracle.SuggestTipCap(ctx)
	gasResult := big.NewInt(0)
	gasResult.Set(tipcap)
	if err != nil {
		return nil, err
	}
	if head := rawdb.ReadCurrentHeader(tx); head != nil && head.BaseFee != nil {
		gasResult.Add(tipcap, head.BaseFee)
	}

	rgp := api.L2GasPircer.GetLastRawGP()
	if gasResult.Cmp(rgp) < 0 {
		gasResult = new(big.Int).Set(rgp)
	}

	return (*hexutil.Big)(gasResult), err
}

func (api *APIImpl) l1GasPrice() (*big.Int, error) {
	res, err := client.JSONRPCCall(api.L1RpcUrl, "eth_gasPrice")
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
