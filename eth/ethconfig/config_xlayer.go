package ethconfig

import (
	"fmt"
	"time"

	"github.com/mitchellh/copystructure"
)

// XLayerConfig is the X Layer config used on the eth backend
type XLayerConfig struct {
	Apollo        ApolloClientConfig
	Nacos         NacosConfig
	EnableInnerTx bool
	// Sequencer
	SequencerBatchSleepDuration time.Duration

	L2Fork9UpgradeBatch uint64
}

var DefaultXLayerConfig = XLayerConfig{}

// NacosConfig is the config for nacos
type NacosConfig struct {
	URLs               string
	NamespaceId        string
	ApplicationName    string
	ExternalListenAddr string
}

// ApolloClientConfig is the config for apollo
type ApolloClientConfig struct {
	Enable        bool
	IP            string
	AppID         string
	NamespaceName string
}

// TryClone is the helper method to return a deep copy of the ethconfig instance
func (c *Config) TryClone() (Config, error) {
	clone, err := copystructure.Copy(*c)
	if err != nil {
		return Config{}, err
	}
	ret, ok := clone.(Config)
	if !ok {
		return Config{}, fmt.Errorf("type assertion failed")
	}
	return ret, nil
}
