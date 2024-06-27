package ethconfig

import "sync"

// ApolloConfig is the apollo eth backend dynamic config
type ApolloConfig struct {
	EnableApollo bool
	conf         Config
	sync.RWMutex
}

var apolloConfig = &ApolloConfig{}

// getApolloConfig returns the singleton instance
func getApolloConfig() *ApolloConfig {
	return apolloConfig
}

func (c *ApolloConfig) get() Config {
	c.RLock()
	defer c.RUnlock()
	return c.conf
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

// UpdateRPCConfig updates the apollo RPC configuration
func UpdateRPCConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add specific gasprice configs to update dynamically
	getApolloConfig().Unlock()
}

// UpdateSequencerConfig updates the apollo sequencer configuration
func UpdateSequencerConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add gasprice configs to update dynamically
	getApolloConfig().Unlock()
}

// UpdatePoolConfig updates the apollo pool configuration
func UpdatePoolConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add pool configs to update dynamically
	getApolloConfig().Unlock()
}

// UpdateL2GasPricerConfig updates the apollo l2gaspricer configuration
func UpdateL2GasPricerConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add l2gaspricer configs to update dynamically
	getApolloConfig().Unlock()
}
