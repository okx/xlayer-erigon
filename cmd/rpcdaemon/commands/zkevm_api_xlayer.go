package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ledgerwatch/erigon/rpc"
	types "github.com/ledgerwatch/erigon/zk/rpcdaemon"
)

func (api *ZkEvmAPIImpl) GetBatchSealTime(ctx context.Context, batchNumberStr string) (types.ArgUint64, error) {
	batchNumber, err := strconv.ParseUint(batchNumberStr, 10, 64)
	if err != nil {
		return 0, err
	}

	lastBatchNo, err := api.BatchNumber(ctx)
	if err != nil {
		return 0, err
	}

	if batchNumber > uint64(lastBatchNo) {
		return 0, errors.New(fmt.Sprintf("couldn't get batch number %d's seal time, error: unexpected batch. got %d, last batch should be %d", batchNumber, batchNumber, lastBatchNo))
	}

	tx, err := api.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	blocks, err := getAllBlocksInBatchNumber(tx, batchNumber)
	if err != nil {
		return 0, err
	}

	if len(blocks) == 0 {
		return 0, errors.New("batch not found")
	}

	lastBlock, err := api.GetFullBlockByNumber(ctx, rpc.BlockNumber(blocks[0]), false)

	return lastBlock.Timestamp, nil
}
