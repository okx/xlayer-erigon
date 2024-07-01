package nodecfg

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

// Enable returns true if apollo is enabled
func (c *ApolloConfig) Enable() bool {
	if c == nil || !c.EnableApollo {
		return false
	}
	c.RLock()
	defer c.RUnlock()
	return c.EnableApollo
}

// UpdateSequencerConfig updates the apollo sequencer configuration
func UpdateSequencerConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add sequencer configs to update dynamically
	getApolloConfig().Unlock()
}

// UpdateRPCConfig updates the apollo RPC configuration
func UpdateRPCConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add specific RPC configs to update dynamically
	getApolloConfig().Unlock()
}

// UpdateL2GasPricerConfig updates the apollo l2gaspricer configuration
func UpdateL2GasPricerConfig(apolloConfig Config) {
	getApolloConfig().Lock()
	getApolloConfig().EnableApollo = true
	// TODO: Add l2gaspricer configs to update dynamically
	getApolloConfig().Unlock()
}
