package apollo

import (
	"testing"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
)

func TestApolloClient_LoadConfig(t *testing.T) {
	c := &ethconfig.Config{
		Zk: &ethconfig.Zk{
			XLayer: ethconfig.XLayerConfig{
				Apollo: ethconfig.ApolloConfig{
					IP:            "",
					AppID:         "xlayer-devnet",
					NamespaceName: "jsonrpc-ro.txt,jsonrpc-roHalt.properties",
					Enable:        true,
				},
			},
		},
	}
	nc := &nodecfg.Config{}
	client := NewClient(c, nc)

	client.LoadConfig()
	t.Log(c.Zk.XLayer.Nacos)
	// time.Sleep(2 * time.Minute)
	t.Log(c.Zk.XLayer.Nacos)
}
