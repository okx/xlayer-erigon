package apollo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
)

func TestLoadJsonRPCConfig(t *testing.T) {

	testFilePath := "../../xlayerconfig-testnet.yaml.example"
	bytes, err := os.ReadFile(testFilePath)
	require.NoError(t, err)
	stringBytes := string(bytes)
	value := interface{}(stringBytes)

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
	client.loadJsonRPC(value)
	require.NoError(t, err)
	logTestNodeConfig(t, client.nodeCfg)
	logTestEthConfig(t, client.ethCfg)
}
