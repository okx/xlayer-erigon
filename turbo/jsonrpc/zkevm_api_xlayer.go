package jsonrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/ledgerwatch/erigon/rpc"
	types "github.com/ledgerwatch/erigon/zk/rpcdaemon"
)

func (api *ZkEvmAPIImpl) GetBatchSealTime(ctx context.Context, batchNumber rpc.BlockNumber) (types.ArgUint64, error) {
	lastBatchNo, err := api.BatchNumber(ctx)
	if err != nil {
		return 0, err
	}

	if batchNumber.Int64() >= int64(lastBatchNo) {
		return 0, errors.New(fmt.Sprintf("couldn't get batch number %d's seal time, error: unexpected batch. got %d, last batch should be %d", batchNumber, batchNumber, lastBatchNo))
	}

	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var lastBlockNum = uint64(0)
	lastBlockNum, err = getLastBlockInBatchNumber(tx, uint64(batchNumber.Int64()))
	if err != nil {
		return 0, err
	}

	lastBlock, err := api.GetFullBlockByNumber(ctx, rpc.BlockNumber(lastBlockNum), false)
	if err != nil {
		return 0, err
	}

	return lastBlock.Timestamp, nil
}
