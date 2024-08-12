package apollo

import (
	"sync"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
)

// Bitset flags for
const (
	JsonRPCFlag = 1 << iota
	SequencerFlag
	L2GasPricerFlag
	PoolFlag
)

// ApolloConfig is the apollo backend dynamic config
type ApolloConfig struct {
	EnableFlag uint32
	NodeCfg    nodecfg.Config
	EthCfg     ethconfig.Config
	sync.RWMutex
}

var apolloConfig = &ApolloConfig{
	EnableFlag: 0,
	NodeCfg:    nodecfg.Config{},
	EthCfg:     ethconfig.Config{},
}

// IsApolloConfigRPCEnabled returns true if the jsonrpc apollo config is enabled
func IsApolloConfigRPCEnabled() bool {
	unsafeGetApolloConfig().RLock()
	defer unsafeGetApolloConfig().RUnlock()
	return unsafeGetApolloConfig().isRPCEnabled()
}

// IsApolloConfigSeqEnabled returns true if the sequencer apollo config is enabled
func IsApolloConfigSequencerEnabled() bool {
	unsafeGetApolloConfig().RLock()
	defer unsafeGetApolloConfig().RUnlock()
	return unsafeGetApolloConfig().isSeqEnabled()
}

// IsApolloConfigGasPricerEnabled returns true if the l2gaspricer apollo config is enabled
func IsApolloConfigL2GasPricerEnabled() bool {
	unsafeGetApolloConfig().RLock()
	defer unsafeGetApolloConfig().RUnlock()
	return unsafeGetApolloConfig().isGPEnabled()
}

// IsApolloConfigPoolEnabled returns true if the pool apollo config is enabled
func IsApolloConfigPoolEnabled() bool {
	unsafeGetApolloConfig().RLock()
	defer unsafeGetApolloConfig().RUnlock()
	return unsafeGetApolloConfig().isPoolEnabled()
}

// unsafeGetApolloConfig is an unsafe function that returns directly the singleton instance
// without locking the sync mutex
// For read operations and most use cases, GetApolloConfig should be used instead
func unsafeGetApolloConfig() *ApolloConfig {
	return apolloConfig
}

// isRPCEnabled returns true if the JsonRPC flag is enabled
func (c *ApolloConfig) isRPCEnabled() bool {
	return c.EnableFlag&JsonRPCFlag != 0
}

// isSeqEnabled returns true if the Sequencer flag is enabled
func (c *ApolloConfig) isSeqEnabled() bool {
	return c.EnableFlag&SequencerFlag != 0
}

// isGPEnabled returns true if the L2GasPricer flag is enabled
func (c *ApolloConfig) isGPEnabled() bool {
	return c.EnableFlag&L2GasPricerFlag != 0
}

// isPoolEnabled returns true if the Pool flag is enabled
func (c *ApolloConfig) isPoolEnabled() bool {
	return c.EnableFlag&PoolFlag != 0
}

func (c *ApolloConfig) setRPCFlag() {
	c.EnableFlag |= JsonRPCFlag
}

func (c *ApolloConfig) setSequencerFlag() {

	c.EnableFlag |= SequencerFlag
}

func (c *ApolloConfig) setGPFlag() {
	c.EnableFlag |= L2GasPricerFlag
}

func (c *ApolloConfig) setPoolFlag() {
	c.EnableFlag |= PoolFlag
}
