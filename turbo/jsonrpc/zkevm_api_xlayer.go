package jsonrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/ledgerwatch/erigon/rpc"
	types "github.com/ledgerwatch/erigon/zk/rpcdaemon"
	"github.com/ledgerwatch/log/v3"
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

func (api *ZkEvmAPIImpl) GetVerifyData(ctx context.Context) ([]string, error) {
	keys := api.ethApi.VerifyCache.Keys()
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		log.Info(fmt.Sprintf("zjg, GetVerifyData, key: %v", key))
		result = append(result, key)
	}
	return result, nil
}

func (api *ZkEvmAPIImpl) SetVerifyData(ctx context.Context, keys []string) (bool, error) {
	if len(keys) == 0 {
		return false, errors.New("no keys provided")
	}
	const p256VerifyInputLength = 160
	for _, key := range keys {
		log.Info(fmt.Sprintf("zjg, key: %s, len:%v", key, len(key)))
		if len(key) == p256VerifyInputLength {
			log.Info("zjg, p256VerifyInputLength")
		}
	}

	return true, nil
}
