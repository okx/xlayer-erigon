package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	proto_txpool "github.com/gateway-fm/cdk-erigon-lib/gointerfaces/txpool"
	"github.com/ledgerwatch/erigon/common/hexutil"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/eth/gasprice"
	"github.com/ledgerwatch/erigon/rpc"
	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
	"github.com/ledgerwatch/log/v3"
)

func (api *APIImpl) gasPriceXL(ctx context.Context) (*hexutil.Big, error) {
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
		return api.gasPriceNonRedirectedXL(ctx)
	}

	price, err := api.getGPFromTrustedNode()
	if err != nil {
		log.Error(fmt.Sprintf("eth_gasPrice error: %v", err))
		return (*hexutil.Big)(api.L2GasPricer.GetConfig().Default), nil
	}

	return (*hexutil.Big)(price), nil
}

func (api *APIImpl) gasPriceNonRedirectedXL(ctx context.Context) (*hexutil.Big, error) {
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

	rgp := api.L2GasPricer.GetLastRawGP()
	if gasResult.Cmp(rgp) < 0 {
		gasResult = new(big.Int).Set(rgp)
	}

	if !api.isCongested(ctx) {
		gasResult = getAvgPrice(rgp, gasResult)
	}

	// For X Layer
	lasthash, _ := api.gasCache.GetLatest()
	api.gasCache.SetLatest(lasthash, gasResult)

	return (*hexutil.Big)(gasResult), err
}

func (api *APIImpl) isCongested(ctx context.Context) bool {

	latestBlockTxNum, err := api.getLatestBlockTxNum(ctx)
	if err != nil {
		return false
	}
	isLatestBlockEmpty := latestBlockTxNum == 0

	poolStatus, err := api.txPool.Status(ctx, &proto_txpool.StatusRequest{})
	if err != nil {
		return false
	}

	isPendingTxCongested := int(poolStatus.PendingCount) >= api.L2GasPricer.GetConfig().XLayer.CongestionThreshold

	return !isLatestBlockEmpty && isPendingTxCongested
}

func (api *APIImpl) getLatestBlockTxNum(ctx context.Context) (int, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	b, err := api.blockByNumber(ctx, rpc.LatestBlockNumber, tx)
	if err != nil {
		return 0, err
	}
	return len(b.Transactions()), nil
}

func (api *APIImpl) getGPFromTrustedNode() (*big.Int, error) {
	res, err := client.JSONRPCCall(api.l2RpcUrl, "eth_gasPrice")
	if err != nil {
		return nil, errors.New("failed to get gas price from trusted node")
	}

	if res.Error != nil {
		return nil, errors.New(res.Error.Message)
	}

	var gasPriceStr string
	err = json.Unmarshal(res.Result, &gasPriceStr)
	if err != nil {
		return nil, errors.New("failed to unmarshal gas price from trusted node")
	}

	gasPriceStr = gasPriceStr[2:]
	gp, err := strconv.ParseUint(gasPriceStr, 16, 64)
	if err != nil {
		return nil, errors.New("failed to parse gas price from trusted node")
	}

	return new(big.Int).SetUint64(gp), nil
}

func (api *APIImpl) runL2GasPriceSuggester() {
	cfg := api.L2GasPricer.GetConfig()
	ctx := api.L2GasPricer.GetCtx()

	//todo: apollo
	l1gp, err := gasprice.GetL1GasPrice(api.L1RpcUrl)
	// if err != nil, do nothing
	if err == nil {
		api.L2GasPricer.UpdateGasPriceAvg(l1gp)
	}
	updateTimer := time.NewTimer(cfg.XLayer.UpdatePeriod)
	for {
		select {
		case <-ctx.Done():
			log.Info("Finishing l2 gas price suggester...")
			return
		case <-updateTimer.C:
			l1gp, err := gasprice.GetL1GasPrice(api.L1RpcUrl)
			if err == nil {
				api.L2GasPricer.UpdateGasPriceAvg(l1gp)
			}
			updateTimer.Reset(cfg.XLayer.UpdatePeriod)
		}
	}
}

func getAvgPrice(low *big.Int, high *big.Int) *big.Int {
	avg := new(big.Int).Add(low, high)
	avg = avg.Quo(avg, big.NewInt(2)) //nolint:gomnd
	return avg
}
