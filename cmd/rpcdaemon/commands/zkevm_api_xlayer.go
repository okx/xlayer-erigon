package commands

import (
	"context"
	"errors"
	"strconv"

	"github.com/ledgerwatch/erigon/rpc"
	types "github.com/ledgerwatch/erigon/zk/rpcdaemon"
)

func (api *ZkEvmAPIImpl) GetBatchSealTime(ctx context.Context, batchNumberStr string) (types.ArgUint64, error) {
	batchNumber, err := strconv.ParseUint(batchNumberStr, 10, 64)
	if err != nil {
		return 0, err
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
