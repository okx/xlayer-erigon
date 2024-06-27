package apollo

import (
	"testing"
	"time"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/stretchr/testify/require"
)

func TestApolloClient_LoadConfig(t *testing.T) {
	c := &ethconfig.Config{
		Zk: &ethconfig.Zk{
			XLayer: ethconfig.XLayerConfig{
				Apollo: ethconfig.ApolloClientConfig{
					IP:            "http://52.40.214.137:26657",
					AppID:         "x1-devnet",
					NamespaceName: "test.txt",
					Enable:        true,
				},
			},
		},
	}
	nc := &nodecfg.Config{}
	client := NewClient(c, nc)

	loaded := client.LoadTestConfig()
	require.Equal(t, true, loaded)

	logTestNodeConfig(t, client.nodeCfg)
	logTestEthConfig(t, client.ethCfg)
	time.Sleep(20 * time.Second)
	logTestNodeConfig(t, client.nodeCfg)
	logTestEthConfig(t, client.ethCfg)
}

func logTestEthConfig(t *testing.T, ethCfg *ethconfig.Config) {
	t.Log("---------- Logging eth backend config ----------")
	t.Log("zkevm.apollo-enable: ", ethCfg.Zk.XLayer.Apollo.Enable)
	t.Log("zkevm.apollo-ip-addr: ", ethCfg.Zk.XLayer.Apollo.IP)
	t.Log("zkevm.apollo-app-id: ", ethCfg.Zk.XLayer.Apollo.AppID)
	t.Log("zkevm.apollo-namespace-name: ", ethCfg.Zk.XLayer.Apollo.NamespaceName)
	t.Log("zkevm.nacos-urls: ", ethCfg.Zk.XLayer.Nacos.URLs)
	t.Log("zkevm.nacos-namespace-id: ", ethCfg.Zk.XLayer.Nacos.NamespaceId)
	t.Log("zkevm.nacos-application-name: ", ethCfg.Zk.XLayer.Nacos.ApplicationName)
	t.Log("zkevm.nacos-external-listen-addr: ", ethCfg.Zk.XLayer.Nacos.ExternalListenAddr)
	t.Log("zkevm.l1-rollup-id", ethCfg.Zk.L1RollupId)
	t.Log("zkevm.l1-first-block", ethCfg.Zk.L1FirstBlock)
	t.Log("zkevm.l1-block-range", ethCfg.Zk.L1BlockRange)
	t.Log("zkevm.l1-query-delay", ethCfg.Zk.L1QueryDelay)
}

func logTestNodeConfig(t *testing.T, nodeCfg *nodecfg.Config) {
	t.Log("---------- Logging node config ----------")
	t.Log("http.addr: ", nodeCfg.Http.HttpListenAddress)
	t.Log("http.port: ", nodeCfg.Http.HttpPort)
	t.Log("http.api: ", nodeCfg.Http.API)
	t.Log("http.timeouts.read: ", nodeCfg.Http.HTTPTimeouts.ReadTimeout)
	t.Log("http.timeouts.read: ", nodeCfg.Http.HTTPTimeouts.WriteTimeout)
	t.Log("ws: ", nodeCfg.Http.WebsocketEnabled)
}
