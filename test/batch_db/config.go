package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DataStreamPort    uint16 `yaml:"zkevm.data-stream-port"`
	DatastreamVersion uint8  `yaml:"zkevm.datastream-version"`
	L2ChainID         uint64 `yaml:"zkevm.l2-chain-id"`
	DatastreamFile    string `yaml:"zkevm.datastream-file"`
	BatchDBPath       string `yaml:"zkevm.batch-db-path"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}
