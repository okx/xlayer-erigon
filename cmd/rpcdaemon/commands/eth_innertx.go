package commands

import (
	"context"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/vm"
	"github.com/ledgerwatch/erigon/eth/stagedsync/stages"
	"github.com/ledgerwatch/erigon/rlp"
	"github.com/ledgerwatch/erigon/rpc"
	"github.com/ledgerwatch/erigon/turbo/rpchelper"
	"github.com/ledgerwatch/erigon/zk/hermez_db"
)

// GetInternalTransactions ...
func (api *APIImpl) GetInternalTransactions(ctx context.Context, txnHash libcommon.Hash) ([]*vm.InnerTx, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	data, err := hermez_db.NewHermezDbReader(tx).GetInnerTxs(txnHash)
	if err != nil {
		return nil, err
	}

	innerTxs := make([]*vm.InnerTx, 0)
	err = rlp.DecodeBytes(data, &innerTxs)
	if err != nil {
		return nil, err
	}
	return innerTxs, nil
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
	//body, err := api._blockReader.BodyWithTransactions(ctx, tx, blockHash, blockNum)
	//if err != nil {
	//	return nil, err
	//}
	//
	//var rtn = make(map[libcommon.Hash][]*vm.InnerTx)
	//for _, txn := range body.Transactions {
	//	data, err := hermez_db.NewHermezDbReader(tx).GetInnerTxs(txn.Hash())
	//	if err != nil {
	//		return nil, err
	//	}
	//	innerTxs := make([]*vm.InnerTx, 0)
	//	err = rlp.DecodeBytes(data, &innerTxs)
	//	if err != nil {
	//		return nil, err
	//	}
	//	rtn[txn.Hash()] = innerTxs
	//}
	//return rtn, nil
}
