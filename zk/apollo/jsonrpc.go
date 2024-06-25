package apollo

import "github.com/apolloconfig/agollo/v4/storage"

// fireJsonRPC fires the json-rpc config change
// BatchRequestsEnabled
// BatchRequestsLimit
// GasLimitFactor
// DisableAPIs
func (c *Client) fireJsonRPC(key string, value *storage.ConfigChange) {
	// newConf, err := c.unmarshal(value.NewValue)
	// if err != nil {
	// 	log.Errorf("failed to unmarshal json-rpc config: %v error: %v", value.NewValue, err)
	// 	return
	// }
	// log.Infof("apollo json-rpc old config : %+v", value.OldValue.(string))
	// log.Infof("apollo json-rpc config changed: %+v", value.NewValue.(string))
	// jsonrpc.UpdateConfig(newConf.RPC)
}
