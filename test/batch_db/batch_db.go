package main

import (
	"encoding/binary"
	"fmt"

	"github.com/erigontech/mdbx-go/mdbx"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/zk/datastream/types"
	"google.golang.org/protobuf/proto"
)

const (
	BatchDBName = "batches"
)

type BatchDB struct {
	env  *mdbx.Env
	path string
	dbi  mdbx.DBI
}

func NewBatchDB(path string) (*BatchDB, error) {
	env, err := mdbx.NewEnv()
	if err != nil {
		return nil, fmt.Errorf("create env: %w", err)
	}

	// 设置最大数据库数量为 1
	if err := env.SetOption(mdbx.OptMaxDB, 1); err != nil {
		return nil, fmt.Errorf("set max dbs: %w", err)
	}

	flags := uint(mdbx.NoTLS | mdbx.NoReadahead | mdbx.WriteMap)
	// 打开/创建数据库文件
	if err := env.Open(path, flags, 0644); err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	// 初始化数据库
	txn, err := env.BeginTxn(nil, 0)
	if err != nil {
		return nil, fmt.Errorf("begin txn: %w", err)
	}
	defer txn.Abort()

	// 打开/创建数据库
	dbi, err := txn.OpenDBI(BatchDBName, mdbx.Create, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("open dbi: %w", err)
	}

	if _, err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &BatchDB{
		env:  env,
		path: path,
		dbi:  dbi,
	}, nil
}

func (db *BatchDB) Close() error {
	db.env.CloseDBI(db.dbi)
	db.env.Close()
	return nil
}

func convertToProtoBlock(block *types.FullL2Block) *BatchData_FullL2Block {
	return &BatchData_FullL2Block{
		BatchNumber:     block.BatchNumber,
		L2BlockNumber:   block.L2BlockNumber,
		Timestamp:       block.Timestamp,
		DeltaTimestamp:  block.DeltaTimestamp,
		L1InfoTreeIndex: block.L1InfoTreeIndex,
		GlobalExitRoot:  block.GlobalExitRoot.Bytes(),
		Coinbase:        block.Coinbase.Bytes(),
		ForkId:          block.ForkId,
		L1BlockHash:     block.L1BlockHash.Bytes(),
		L2BlockHash:     block.L2Blockhash.Bytes(),
		ParentHash:      block.ParentHash.Bytes(),
		StateRoot:       block.StateRoot.Bytes(),
		BlockGasLimit:   block.BlockGasLimit,
		BlockInfoRoot:   block.BlockInfoRoot.Bytes(),
	}
}

func convertFromProtoBlock(block *BatchData_FullL2Block) *types.FullL2Block {
	return &types.FullL2Block{
		BatchNumber:     block.BatchNumber,
		L2BlockNumber:   block.L2BlockNumber,
		Timestamp:       block.Timestamp,
		DeltaTimestamp:  block.DeltaTimestamp,
		L1InfoTreeIndex: block.L1InfoTreeIndex,
		GlobalExitRoot:  libcommon.BytesToHash(block.GlobalExitRoot),
		Coinbase:        libcommon.BytesToAddress(block.Coinbase),
		ForkId:          block.ForkId,
		L1BlockHash:     libcommon.BytesToHash(block.L1BlockHash),
		L2Blockhash:     libcommon.BytesToHash(block.L2BlockHash),
		ParentHash:      libcommon.BytesToHash(block.ParentHash),
		StateRoot:       libcommon.BytesToHash(block.StateRoot),
		BlockGasLimit:   block.BlockGasLimit,
		BlockInfoRoot:   libcommon.BytesToHash(block.BlockInfoRoot),
	}
}

func (db *BatchDB) StoreBatch(batch []*types.FullL2Block) error {
	if len(batch) == 0 {
		return nil
	}

	protoBlocks := make([]*BatchData_FullL2Block, len(batch))
	for i, block := range batch {
		protoBlocks[i] = convertToProtoBlock(block)
	}

	batchData := &BatchData{
		Blocks: protoBlocks,
	}

	data, err := proto.Marshal(batchData)
	if err != nil {
		return fmt.Errorf("marshal batch %d: %w", batch[0].BatchNumber, err)
	}

	return db.env.Update(func(txn *mdbx.Txn) error {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, batch[0].BatchNumber)
		return txn.Put(db.dbi, key, data, 0)
	})
}

func (db *BatchDB) StoreBatches(batches [][]*types.FullL2Block) error {
	if len(batches) == 0 {
		return nil
	}

	// 使用事务批量写入
	return db.env.Update(func(txn *mdbx.Txn) error {
		for _, batch := range batches {
			if len(batch) == 0 {
				continue
			}

			// 转换为 proto 格式
			protoBlocks := make([]*BatchData_FullL2Block, len(batch))
			for i, block := range batch {
				protoBlocks[i] = convertToProtoBlock(block)
			}

			batchData := &BatchData{
				Blocks: protoBlocks,
			}

			// 序列化数据
			data, err := proto.Marshal(batchData)
			if err != nil {
				return fmt.Errorf("marshal batch %d: %w", batch[0].BatchNumber, err)
			}

			// 存储到数据库
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, batch[0].BatchNumber)
			if err := txn.Put(db.dbi, key, data, 0); err != nil {
				return fmt.Errorf("store batch %d: %w", batch[0].BatchNumber, err)
			}
		}
		return nil
	})
}

func (db *BatchDB) GetBatch(batchNumber uint64) ([]*types.FullL2Block, error) {
	var data []byte
	err := db.env.View(func(txn *mdbx.Txn) error {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, batchNumber)

		val, err := txn.Get(db.dbi, key)
		if err != nil {
			return err
		}
		data = val
		return nil
	})
	if err != nil {
		return nil, err
	}

	var batchData BatchData
	if err := proto.Unmarshal(data, &batchData); err != nil {
		return nil, fmt.Errorf("unmarshal batch %d: %w", batchNumber, err)
	}

	result := make([]*types.FullL2Block, len(batchData.Blocks))
	for i, block := range batchData.Blocks {
		result[i] = convertFromProtoBlock(block)
	}

	return result, nil
}

func (db *BatchDB) GetBatchRange(fromBatch, toBatch uint64) ([][]*types.FullL2Block, error) {
	var result [][]*types.FullL2Block

	err := db.env.View(func(txn *mdbx.Txn) error {
		cursor, err := txn.OpenCursor(db.dbi)
		if err != nil {
			return err
		}
		defer cursor.Close()

		startKey := make([]byte, 8)
		binary.BigEndian.PutUint64(startKey, fromBatch)

		for key, val, err := cursor.Get(startKey, nil, mdbx.SetRange); err == nil; key, val, err = cursor.Get(nil, nil, mdbx.Next) {
			if key == nil {
				break
			}

			batchNum := binary.BigEndian.Uint64(key)
			if batchNum > toBatch {
				break
			}

			var batchData BatchData
			if err := proto.Unmarshal(val, &batchData); err != nil {
				return fmt.Errorf("unmarshal batch %d: %w", batchNum, err)
			}

			blocks := make([]*types.FullL2Block, len(batchData.Blocks))
			for i, block := range batchData.Blocks {
				blocks[i] = convertFromProtoBlock(block)
			}
			result = append(result, blocks)
		}
		return nil
	})

	return result, err
}
