package commands

import (
	"context"

	libcommon "github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/core/vm"
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
func (api *APIImpl) GetBlockInternalTransactions(ctx context.Context, numberOrHash rpc.BlockNumberOrHash) (map[libcommon.Hash][]*vm.InnerTx, error) {
	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	blockNum, blockHash, _, err := rpchelper.GetBlockNumber(numberOrHash, tx, api.filters)
	if err != nil {
		return nil, err
	}
	if int64(blockNum) == rpc.PendingBlockNumber.Int64() {
		return nil, nil
	}
	latestBlockNumber, err := rpchelper.GetLatestBlockNumber(tx)
	if err != nil {
		return nil, err
	}
	if blockNum > latestBlockNumber {
		return nil, nil
	}

	body, err := api._blockReader.BodyWithTransactions(ctx, tx, blockHash, blockNum)
	if err != nil {
		return nil, err
	}

	var rtn = make(map[libcommon.Hash][]*vm.InnerTx)
	for _, txn := range body.Transactions {
		data, err := hermez_db.NewHermezDbReader(tx).GetInnerTxs(txn.Hash())
		if err != nil {
			return nil, err
		}
		innerTxs := make([]*vm.InnerTx, 0)
		err = rlp.DecodeBytes(data, &innerTxs)
		if err != nil {
			return nil, err
		}
		rtn[txn.Hash()] = innerTxs
	}
	return rtn, nil
}
