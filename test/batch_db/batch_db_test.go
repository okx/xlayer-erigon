package main

import (
	"os"
	"path/filepath"
	"testing"

	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/zk/datastream/types"
	"github.com/stretchr/testify/assert"
)

func createTestBlock(batchNum uint64, blockNum uint64) *types.FullL2Block {
	return &types.FullL2Block{
		BatchNumber:     batchNum,
		L2BlockNumber:   blockNum,
		Timestamp:       1234567890,
		DeltaTimestamp:  10,
		L1InfoTreeIndex: 1,
		GlobalExitRoot:  libcommon.HexToHash("0x1234"),
		Coinbase:        libcommon.HexToAddress("0x5678"),
		ForkId:          1,
		L1BlockHash:     libcommon.HexToHash("0x9abc"),
		L2Blockhash:     libcommon.HexToHash("0xdef0"),
		ParentHash:      libcommon.HexToHash("0x1111"),
		StateRoot:       libcommon.HexToHash("0x2222"),
		BlockGasLimit:   1000000,
		BlockInfoRoot:   libcommon.HexToHash("0x3333"),
	}
}

func TestBatchDB(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir, err := os.MkdirTemp("", "batch_db_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.mdbx")

	// 测试创建数据库
	db, err := NewBatchDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 测试存储批次
	batch1 := []*types.FullL2Block{
		createTestBlock(1, 1),
		createTestBlock(1, 2),
	}
	batch2 := []*types.FullL2Block{
		createTestBlock(2, 3),
		createTestBlock(2, 4),
	}

	err = db.StoreBatch(batch1)
	assert.NoError(t, err)
	err = db.StoreBatch(batch2)
	assert.NoError(t, err)

	// 测试获取单个批次
	blocks, err := db.GetBatch(1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(blocks))
	assert.Equal(t, uint64(1), blocks[0].BatchNumber)
	assert.Equal(t, uint64(1), blocks[0].L2BlockNumber)
	assert.Equal(t, uint64(2), blocks[1].L2BlockNumber)

	// 测试获取批次范围
	batches, err := db.GetBatchRange(1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(batches))
	assert.Equal(t, 2, len(batches[0]))
	assert.Equal(t, 2, len(batches[1]))
	assert.Equal(t, uint64(1), batches[0][0].BatchNumber)
	assert.Equal(t, uint64(2), batches[1][0].BatchNumber)

	// 测试空批次
	err = db.StoreBatch(nil)
	assert.NoError(t, err)

	// 测试不存在的批次
	blocks, err = db.GetBatch(999)
	assert.Error(t, err)

	// 测试范围查询边界情况
	batches, err = db.GetBatchRange(999, 1000)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(batches))
}

func TestBatchDB_StoreBatches(t *testing.T) {
	// 创建临时目录用于测试
	tmpDir, err := os.MkdirTemp("", "batch_db_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.mdbx")

	// 创建数据库实例
	db, err := NewBatchDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tests := []struct {
		name    string
		batches [][]*types.FullL2Block
		wantErr bool
	}{
		{
			name:    "空批次列表",
			batches: nil,
			wantErr: false,
		},
		{
			name: "包含空批次",
			batches: [][]*types.FullL2Block{
				{},
				{
					createTestBlock(1, 1),
				},
				{},
			},
			wantErr: false,
		},
		{
			name: "多个正常批次",
			batches: [][]*types.FullL2Block{
				{
					createTestBlock(1, 1),
					createTestBlock(1, 2),
				},
				{
					createTestBlock(2, 3),
					createTestBlock(2, 4),
				},
			},
			wantErr: false,
		},
		{
			name: "批次号连续性测试",
			batches: [][]*types.FullL2Block{
				{createTestBlock(5, 1)},
				{createTestBlock(6, 2)},
				{createTestBlock(7, 3)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 存储批次
			err := db.StoreBatches(tt.batches)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 验证存储结果
			if tt.batches != nil {
				for _, batch := range tt.batches {
					if len(batch) == 0 {
						continue
					}

					// 读取并验证每个批次
					batchNum := batch[0].BatchNumber
					stored, err := db.GetBatch(batchNum)
					assert.NoError(t, err)
					assert.Equal(t, len(batch), len(stored))

					// 验证每个区块的内容
					for i, block := range batch {
						assert.Equal(t, block.BatchNumber, stored[i].BatchNumber)
						assert.Equal(t, block.L2BlockNumber, stored[i].L2BlockNumber)
						assert.Equal(t, block.Timestamp, stored[i].Timestamp)
						assert.Equal(t, block.GlobalExitRoot, stored[i].GlobalExitRoot)
						assert.Equal(t, block.Coinbase, stored[i].Coinbase)
						assert.Equal(t, block.L1BlockHash, stored[i].L1BlockHash)
						assert.Equal(t, block.L2Blockhash, stored[i].L2Blockhash)
						assert.Equal(t, block.StateRoot, stored[i].StateRoot)
					}
				}
			}

			// 验证范围查询
			if tt.batches != nil && len(tt.batches) > 0 {
				var minBatch, maxBatch uint64 = ^uint64(0), 0
				for _, batch := range tt.batches {
					if len(batch) > 0 {
						batchNum := batch[0].BatchNumber
						if batchNum < minBatch {
							minBatch = batchNum
						}
						if batchNum > maxBatch {
							maxBatch = batchNum
						}
					}
				}
				if minBatch != ^uint64(0) {
					batches, err := db.GetBatchRange(minBatch, maxBatch)
					assert.NoError(t, err)
					assert.NotEmpty(t, batches)
				}
			}
		})
	}
}
