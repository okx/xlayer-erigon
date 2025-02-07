package main

import (
	"fmt"
	"time"

	"github.com/0xPolygonHermez/zkevm-data-streamer/datastreamer"
	"github.com/0xPolygonHermez/zkevm-data-streamer/log"
	"github.com/ledgerwatch/erigon/zk/datastream/server"
	"github.com/ledgerwatch/erigon/zk/datastream/types"
)

func PrintBlockInfo(block *types.FullL2Block) {
	fmt.Printf("  Block %d:\n", block.L2BlockNumber)
	fmt.Printf("    BatchNumber: %d\n", block.BatchNumber)
	fmt.Printf("    L2BlockNumber: %d\n", block.L2BlockNumber)
	fmt.Printf("    Timestamp: %d\n", block.Timestamp)
	fmt.Printf("    DeltaTimestamp: %d\n", block.DeltaTimestamp)
	fmt.Printf("    L1InfoTreeIndex: %d\n", block.L1InfoTreeIndex)
	fmt.Printf("    GlobalExitRoot: %s\n", block.GlobalExitRoot)
	fmt.Printf("    Coinbase: %s\n", block.Coinbase)
	fmt.Printf("    ForkId: %d\n", block.ForkId)
	fmt.Printf("    L1BlockHash: %s\n", block.L1BlockHash)
	fmt.Printf("    L2BlockHash: %s\n", block.L2Blockhash)
	fmt.Printf("    ParentHash: %s\n", block.ParentHash)
	fmt.Printf("    StateRoot: %s\n", block.StateRoot)
	fmt.Printf("    BlockGasLimit: %d\n", block.BlockGasLimit)
	fmt.Printf("    BlockInfoRoot: %s\n", block.BlockInfoRoot)
	fmt.Printf("    Transactions count: %d\n", len(block.L2Txs))
	fmt.Println()
}

func PrintBatchInfo(batch []*types.FullL2Block) {
	if len(batch) == 0 {
		return
	}
	fmt.Printf("Batch number %d:\n", batch[0].BatchNumber)
	for _, block := range batch {
		PrintBlockInfo(block)
	}
	fmt.Println("------------------------")
}

func PrintBatchesInfo(batches [][]*types.FullL2Block) {
	for _, batch := range batches {
		PrintBatchInfo(batch)
	}
}

// 将函数名改为大写开头
func ReadDataStreamBatches(config *Config, fromBatch, toBatch uint64) ([][]*types.FullL2Block, error) {
	// 函数内部逻辑保持不变
	// Use hardcoded timeout values
	writeTimeout := 20 * time.Second
	inactivityTimeout := 10 * time.Minute
	inactivityCheckInterval := 5 * time.Minute

	// 设置日志配置
	logConfig := &log.Config{
		Environment: "production",
		Level:       "warn",
		Outputs:     nil,
	}

	// 创建 stream server factory
	factory := server.NewZkEVMDataStreamServerFactory()

	// 创建 stream server
	streamServer, err := factory.CreateStreamServer(
		config.DataStreamPort,
		config.DatastreamVersion,
		1,
		datastreamer.StreamType(1),
		config.DatastreamFile,
		writeTimeout,
		inactivityTimeout,
		inactivityCheckInterval,
		logConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream server: %v", err)
	}

	fmt.Printf("Successfully created stream server with file: %s\n", config.DatastreamFile)

	// 创建 data stream server
	dataStreamServer := factory.CreateDataStreamServer(streamServer, config.L2ChainID)

	// 获取数据流中的最高批次号
	highestBatchInDs, err := dataStreamServer.GetHighestBatchNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to get highest batch number: %v", err)
	}

	// 确保 toBatch 不超过最高批次号
	if toBatch > highestBatchInDs {
		fmt.Printf("Adjusting toBatch from %d to %d (highest batch in datastream)\n", toBatch, highestBatchInDs)
		toBatch = highestBatchInDs
	}

	// 读取 batches
	batches, err := dataStreamServer.ReadBatches(fromBatch, toBatch)
	if err != nil {
		return nil, fmt.Errorf("failed to read batches: %v", err)
	}

	return batches, nil
}

// func main() {
// 	// Load configuration from YAML file
// 	configPath := "test.datastream.config.yaml"
// 	config, err := LoadConfig(configPath)
// 	if err != nil {
// 		log.Fatalf("Failed to load config: %v", err)
// 	}

// 	// 读取 batches
// 	fromBatch := uint64(1) // TODO: 填入起始 batch 号
// 	toBatch := uint64(10)  // TODO: 填入结束 batch 号

// 	batches, err := readDataStreamBatches(config, fromBatch, toBatch)
// 	if err != nil {
// 		log.Fatalf("Failed to read data stream batches: %v", err)
// 	}

// 	// Print batch information
// 	printBatchesInfo(batches)
// }
