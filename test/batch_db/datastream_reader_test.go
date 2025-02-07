package main

import (
	"testing"
)

func TestDataStreamReader(t *testing.T) {
	// Load configuration from YAML file
	configPath := "test.datastream.config.yaml"
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 读取 batches
	fromBatch := uint64(1)
	toBatch := uint64(10)

	batches, err := ReadDataStreamBatches(config, fromBatch, toBatch)
	if err != nil {
		t.Fatalf("Failed to read data stream batches: %v", err)
	}

	// 验证读取的数据
	if len(batches) == 0 {
		t.Error("No batches were read")
	}

	// 打印批次信息用于调试
	t.Log("Read batches successfully:")
	for i, batch := range batches {
		if len(batch) == 0 {
			t.Errorf("Batch %d is empty", i)
			continue
		}
		t.Logf("Batch %d contains %d blocks", batch[0].BatchNumber, len(batch))
	}
}
