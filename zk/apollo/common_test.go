package apollo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
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
					NamespaceName: "jsonrpc-ro.txt, sequencer-roHalt.properties",
					Enable:        true,
				},
			},
		},
	}
	client := NewClient(c)
	client.loadJsonRPC(value)
	require.NoError(t, err)

	apolloNodeCfg := UnsafeGetApolloConfig().NodeCfg
	apolloEthCfg := UnsafeGetApolloConfig().EthCfg

	logTestNodeConfig(t, &apolloNodeCfg)
	logTestEthConfig(t, &apolloEthCfg)
}
