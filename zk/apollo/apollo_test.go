package apollo

import (
	"fmt"
	"testing"
	"time"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/stretchr/testify/require"
)

func TestApolloClient(t *testing.T) {
	c := &ethconfig.Config{
		Zk: &ethconfig.Zk{
			XLayer: ethconfig.XLayerConfig{
				Apollo: ethconfig.ApolloClientConfig{
					IP:            "http://127.0.0.1:18080",
					AppID:         "SampleApp",
					NamespaceName: "jsonrpc-tester.txt",
					Enable:        true,
				},
			},
		},
	}
	client := NewClient(c)

	// Test load config cache
	loaded := client.LoadConfig()
	require.Equal(t, true, loaded)

	apolloInitNodeCfg, err := nodecfg.GetApolloConfig()
	require.NoError(t, err)
	apolloInitEthCfg, err := ethconfig.GetApolloConfig()
	require.NoError(t, err)

	logTestNodeConfig(t, &apolloInitNodeCfg)
	logTestEthConfig(t, &apolloInitEthCfg)

	// Fire config changes on both ethconfig and nodecfg
	time.Sleep(30 * time.Second)

	apolloAfterNodeCfg, err := nodecfg.GetApolloConfig()
	require.NoError(t, err)
	apolloAfterEthCfg, err := ethconfig.GetApolloConfig()
	require.NoError(t, err)

	t.Log("Logging apollo config")
	logTestNodeConfig(t, &apolloAfterNodeCfg)
	logTestEthConfig(t, &apolloAfterEthCfg)

	// Test apollo dynamic configuration changes
	require.NotEqual(t, apolloInitNodeCfg, apolloAfterNodeCfg)
	require.NotEqual(t, apolloInitEthCfg, apolloAfterEthCfg)
}

func TestApolloConfig(t *testing.T) {
	c := &ethconfig.Config{
		Zk: &ethconfig.Zk{
			XLayer: ethconfig.XLayerConfig{
				Apollo: ethconfig.ApolloClientConfig{
					IP:            "http://127.0.0.1:18080",
					AppID:         "SampleApp",
					NamespaceName: "sequencer-hihihi.txt",
					Enable:        true,
				},
			},
		},
	}
	client := NewClient(c)

	// Test load config cache
	loaded := client.LoadConfig()
	require.Equal(t, true, loaded)

	// Test to ensure read apollo nodecfg method return deep copies
	firstCopyApolloNodeCfg, err := nodecfg.GetApolloConfig()
	require.NoError(t, err)
	secondCopyApolloNodeCfg, err := nodecfg.GetApolloConfig()
	require.NoError(t, err)
	t.Logf("1st copy address of node config: %p", &firstCopyApolloNodeCfg)
	t.Logf("2nd copy address of node config: %p", &secondCopyApolloNodeCfg)
	require.NotEqual(t, fmt.Sprintf("%p", &firstCopyApolloNodeCfg), fmt.Sprintf("%p", &secondCopyApolloNodeCfg))
	require.Equal(t, firstCopyApolloNodeCfg, secondCopyApolloNodeCfg)

	// Test to ensure read apollo ethconfig method return deep copies
	firstCopyApolloEthCfg, err := ethconfig.GetApolloConfig()
	require.NoError(t, err)
	secondCopyApolloEthCfg, err := ethconfig.GetApolloConfig()
	require.NoError(t, err)
	t.Logf("1st copy address of ethconfig: %p", &firstCopyApolloEthCfg)
	t.Logf("2nd copy address of ethconfig: %p", &secondCopyApolloEthCfg)
	require.NotEqual(t, fmt.Sprintf("%p", &firstCopyApolloEthCfg), fmt.Sprintf("%p", &secondCopyApolloEthCfg))
	t.Logf("1st copy address of zk config: %p", firstCopyApolloEthCfg.Zk)
	t.Logf("2nd copy address of zk config: %p", secondCopyApolloEthCfg.Zk)
	require.NotEqual(t, fmt.Sprintf("%p", firstCopyApolloEthCfg.Zk), fmt.Sprintf("%p", secondCopyApolloEthCfg.Zk))
	require.Equal(t, firstCopyApolloEthCfg, secondCopyApolloEthCfg)
}

func logTestEthConfig(t *testing.T, ethCfg *ethconfig.Config) {
	t.Log("---------- Logging eth backend config ----------")
	t.Log("zkevm.apollo-enabled: ", ethCfg.Zk.XLayer.Apollo.Enable)
	t.Log("zkevm.apollo-ip-addr: ", ethCfg.Zk.XLayer.Apollo.IP)
	t.Log("zkevm.apollo-app-id: ", ethCfg.Zk.XLayer.Apollo.AppID)
	t.Log("zkevm.apollo-namespace-name: ", ethCfg.Zk.XLayer.Apollo.NamespaceName)
	t.Log("zkevm.nacos-urls: ", ethCfg.Zk.XLayer.Nacos.URLs)
	t.Log("zkevm.nacos-namespace-id: ", ethCfg.Zk.XLayer.Nacos.NamespaceId)
	t.Log("zkevm.nacos-application-name: ", ethCfg.Zk.XLayer.Nacos.ApplicationName)
	t.Log("zkevm.nacos-external-listen-addr: ", ethCfg.Zk.XLayer.Nacos.ExternalListenAddr)
	t.Log("zkevm.l1-rollup-id: ", ethCfg.Zk.L1RollupId)
	t.Log("zkevm.l1-first-block: ", ethCfg.Zk.L1FirstBlock)
	t.Log("zkevm.l1-block-range: ", ethCfg.Zk.L1BlockRange)
	t.Log("zkevm.l1-query-delay: ", ethCfg.Zk.L1QueryDelay)
	t.Log("zkevm.sequencer-block-seal-time: ", ethCfg.Zk.SequencerBlockSealTime)
	t.Log("zkevm.sequencer-batch-seal-time: ", ethCfg.Zk.SequencerBatchSealTime)
	t.Log("zkevm.sequencer-non-empty-batch-seal-time: ", ethCfg.Zk.SequencerNonEmptyBatchSealTime)
}

func logTestNodeConfig(t *testing.T, nodeCfg *nodecfg.Config) {
	t.Log("---------- Logging node config ----------")
	t.Log("http.addr: ", nodeCfg.Http.HttpListenAddress)
	t.Log("http.port: ", nodeCfg.Http.HttpPort)
	t.Log("http.api: ", nodeCfg.Http.API)
	t.Log("http.timeouts.read: ", nodeCfg.Http.HTTPTimeouts.ReadTimeout)
	t.Log("http.timeouts.write: ", nodeCfg.Http.HTTPTimeouts.WriteTimeout)
	t.Log("ws: ", nodeCfg.Http.WebsocketEnabled)
}
