package main

import (
	"fmt"

	"github.com/0xPolygonHermez/zkevm-data-streamer/log"
)

func main() {
	// 加载配置文件
	configPath := "test.datastream.config.yaml" // 修改配置文件路径
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 BatchDB 实例
	db, err := NewBatchDB(config.BatchDBPath)
	if err != nil {
		log.Fatalf("Failed to create batch db: %v", err)
	}
	defer db.Close()

	// 读取 datastream 中的 batches
	fromBatch := uint64(1)
	toBatch := uint64(1000)

	batches, err := ReadDataStreamBatches(config, fromBatch, toBatch)
	if err != nil {
		log.Fatalf("Failed to read data stream batches: %v", err)
	}

	// 存储到数据库
	if err := db.StoreBatches(batches); err != nil {
		log.Fatalf("Failed to store batches: %v", err)
	}

	fmt.Printf("Successfully stored %d batches to database\n", len(batches))

	// 打印所有批次的详细信息
	// if len(batches) > 0 {
	// 	fmt.Println("\nAll batches details:")
	// 	PrintBatchesInfo(batches)
	// }

	// 验证存储结果
	storedBatches, err := db.GetBatchRange(fromBatch, toBatch)
	if err != nil {
		log.Fatalf("Failed to verify stored batches: %v", err)
	}

	fmt.Printf("Successfully verified %d batches in database\n", len(storedBatches))
}
