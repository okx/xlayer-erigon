package hermez_db

import (
	"encoding/binary"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/dbutils"
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
	defer func() {
		if casted, ok := it.(kv.Closer); ok {
			casted.Close()
		}
	}()

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

// TruncateInnerTx deletes all inner txs of a block
func (db *HermezDb) TruncateInnerTx(block uint64) error {
	prefix := make([]byte, 8)
	binary.BigEndian.PutUint64(prefix, block)

	it, err := db.tx.Prefix(INNER_TX, prefix)
	if err != nil {
		log.Error("inner txs fetching failed", "err", err)
		return nil
	}
	defer func() {
		if casted, ok := it.(kv.Closer); ok {
			casted.Close()
		}
	}()

	var keyList [][]byte
	for it.HasNext() {
		k, v, err := it.Next()
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
		keyCopy := make([]byte, len(k))
		copy(keyCopy, k)
		keyList = append(keyList, keyCopy)
	}

	for _, k := range keyList {
		err = db.tx.Delete(INNER_TX, k)
		if err != nil {
			log.Error("inner txs fetching failed", "err", err)
			return err
		}
	}

	afterTxs := db.GetInnerTxs(block)
	afterCount := len(afterTxs)
	log.Info("Delete inner txs", "block", block, "delete count", len(keyList), "after count", afterCount)
	return nil
}
