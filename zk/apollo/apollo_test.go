package apollo

import (
	"testing"
	"time"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
)

func TestApolloClient_LoadConfig(t *testing.T) {
	c := &ethconfig.Config{
		Zk: &ethconfig.Zk{
			XLayer: ethconfig.XLayerConfig{
				Apollo: ethconfig.ApolloClientConfig{
					IP:            "0.0.0.0",
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
	t.Log(c.Zk.XLayer)
	time.Sleep(2 * time.Minute)
	t.Log(c.Zk.XLayer)
}
