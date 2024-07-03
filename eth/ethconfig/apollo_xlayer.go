package ethconfig

import (
	"fmt"
	"sync"
)

// ApolloConfig is the apollo eth backend dynamic config
type ApolloConfig struct {
	EnableApollo bool
	Conf         Config
	sync.RWMutex
}

var apolloConfig = &ApolloConfig{}

// GetApolloConfig returns a copy of the singleton instance apollo config
func GetApolloConfig() (Config, error) {
	if UnsafeGetApolloConfig().Enable() {
		UnsafeGetApolloConfig().RLock()
		defer UnsafeGetApolloConfig().RUnlock()
		return UnsafeGetApolloConfig().Conf, nil
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
