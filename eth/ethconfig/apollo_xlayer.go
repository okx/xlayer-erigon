package ethconfig

import (
	"fmt"
	"sync"
	"time"

	"github.com/ledgerwatch/erigon/zkevm/log"
)

// ApolloConfig is the apollo eth backend dynamic config
type ApolloConfig struct {
	EnableApollo bool
	Conf         Config
	sync.RWMutex
}

var apolloConfig = &ApolloConfig{
	EnableApollo: false,
	Conf: Config{
		Zk: &Zk{},
	},
}

// GetApolloConfig returns a copy of the singleton instance apollo config
func GetApolloConfig() (Config, error) {
	if UnsafeGetApolloConfig().Enable() {
		UnsafeGetApolloConfig().RLock()
		defer UnsafeGetApolloConfig().RUnlock()
		conf, err := UnsafeGetApolloConfig().Conf.TryClone()
		if err != nil {
			log.Debug(fmt.Sprintf("GetApolloConfig error: %v", err))
			return Config{}, err
		}
		return conf, nil
	} else {
		return Config{}, fmt.Errorf("apollo config disabled")
	}
}

// UnsafeGetApolloConfig is an unsafe function that returns directly the singleton
// instance without locking the sync mutex
// For read operations and most use cases, GetApolloConfig should be used instead
func UnsafeGetApolloConfig() *ApolloConfig {
	return apolloConfig
}

// Enable returns true if apollo is enabled
func (c *ApolloConfig) Enable() bool {
	if c == nil || !c.EnableApollo {
		return false
	}
	c.RLock()
	defer c.RUnlock()
	return c.EnableApollo
}

func GetFullBatchSleepDuration(localDuration time.Duration) time.Duration {
	conf, err := GetApolloConfig()
	if err != nil {
		return localDuration
	} else {
		return conf.Zk.XLayer.SequencerBatchSleepDuration
	}
}
