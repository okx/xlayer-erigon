package commands

import (
	"context"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/vm"
	"github.com/ledgerwatch/erigon/eth/stagedsync/stages"
	"github.com/ledgerwatch/erigon/rpc"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
)

// GetInternalTransactions ...
func (api *APIImpl) GetInternalTransactions(ctx context.Context, txnHash libcommon.Hash) ([]*vm.InnerTx, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// TODO

	return nil, nil
}

// GetBlockInternalTransactions ...
func (api *APIImpl) GetBlockInternalTransactions(ctx context.Context, number rpc.BlockNumber) ([][]*vm.InnerTx, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if number == rpc.PendingBlockNumber {
		return nil, nil
	}

	// get latest executed block
	executedBlock, err := stages.GetStageProgress(tx, stages.Execution)
	if err != nil {
		return nil, err
	}

	// return null if requested block  is higher than executed
	// made for consistency with zkevm
	if number > 0 && executedBlock < uint64(number.Int64()) {
		return nil, nil
	}

	n, _, _, err := rpchelper.GetBlockNumber(rpc.BlockNumberOrHashWithNumber(number), tx, api.filters)
	if err != nil {
		return nil, err
	}

	return rawdb.ReadInnerTxs(tx, n), nil
}
