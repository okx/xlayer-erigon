package hermez_db

import (
	"encoding/binary"
	"fmt"
	
	"github.com/ledgerwatch/erigon/common/dbutils"
	"github.com/ledgerwatch/erigon/rlp"
	"github.com/ledgerwatch/erigon/zk/types"
	"github.com/ledgerwatch/log/v3"
)

const INNER_TX = "InnerTx" // block_num_u64 + txId -> inner txs of transaction

func (db *HermezDb) WriteInnerTxs(number uint64, innerTxs [][]*types.InnerTx) error {
	for txId, its := range innerTxs {
		if len(its) == 0 {
			continue
		}

		data, err := rlp.EncodeToBytes(its)
		if err != nil {
			return fmt.Errorf("encode inner tx for block %d: %w", number, err)
		}

		if err = db.tx.Append(INNER_TX, dbutils.LogKey(number, uint32(txId)), data); err != nil {
			return fmt.Errorf("writing logs for block %d: %w", number, err)
		}
	}
	return nil
}

func (db *HermezDbReader) GetInnerTxs(blockNum uint64) [][]*types.InnerTx {
	var blockInnerTxs [][]*types.InnerTx

	prefix := make([]byte, 8)
	binary.BigEndian.PutUint64(prefix, blockNum)

	it, err := db.tx.Prefix(INNER_TX, prefix)
	if err != nil {
		log.Error("inner txs fetching failed", "err", err)
		return nil
	}
	for it.HasNext() {
		_, v, err := it.Next()
		if err != nil {
			log.Error("inner txs fetching failed", "err", err)
			return nil
		}

		innerTxs := make([]*types.InnerTx, 0)
		if err := rlp.DecodeBytes(v, &innerTxs); err != nil {
			err = fmt.Errorf("inner txs unmarshal failed:  %w", err)
			log.Error("inner txs fetching failed", "err", err)
			return nil
		}

		blockInnerTxs = append(blockInnerTxs, innerTxs)
	}
	return blockInnerTxs
}
