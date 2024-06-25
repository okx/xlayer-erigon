package commands

import (
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/ledgerwatch/erigon/eth/gasprice"
	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
	"github.com/ledgerwatch/log/v3"
)

func (api *APIImpl) getGPFromTrustedNode() (*big.Int, error) {
	res, err := client.JSONRPCCall(api.l2RpcUrl, "eth_gasPrice")
	if err != nil {
		return nil, errors.New("failed to get gas price from trusted node")
	}

	if res.Error != nil {
		return nil, errors.New(res.Error.Message)
	}

	var gasPrice uint64
	err = json.Unmarshal(res.Result, &gasPrice)
	if err != nil {
		return nil, errors.New("failed to read gas price from trusted node")
	}
	return new(big.Int).SetUint64(gasPrice), nil
}

func (api *APIImpl) runL2GasPriceSuggester() {
	cfg := api.L2GasPircer.GetConfig()
	ctx := api.L2GasPircer.GetCtx()

	//todo: apollo
	l1gp, err := gasprice.GetL1GasPrice(api.L1RpcUrl)
	// if err != nil, do nothing
	if err == nil {
		api.L2GasPircer.UpdateGasPriceAvg(l1gp)
	}
	updateTimer := time.NewTimer(cfg.UpdatePeriod)
	for {
		select {
		case <-ctx.Done():
			log.Info("Finishing l2 gas price suggester...")
			return
		case <-updateTimer.C:
			l1gp, err := gasprice.GetL1GasPrice(api.L1RpcUrl)
			if err == nil {
				api.L2GasPircer.UpdateGasPriceAvg(l1gp)
			}
			updateTimer.Reset(cfg.UpdatePeriod)
		}
	}
}
