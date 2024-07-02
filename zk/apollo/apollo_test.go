package apollo

import (
	"testing"
	"time"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/stretchr/testify/require"
)

func TestJsonRPCApolloClient(t *testing.T) {
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
	nc := &nodecfg.Config{}
	client := NewClient(c, nc)

	loaded := client.LoadConfig()
	require.Equal(t, true, loaded)

	logTestNodeConfig(t, nc)
	t.Log("Logging apollo config")
	apolloCfg := nodecfg.GetApolloConfig()
	logTestNodeConfig(t, &apolloCfg)
	initialHttpConf := nc.Http

	// Fire jsonrpc config changes
	time.Sleep(30 * time.Second)

	afterHttpConf := nodecfg.GetApolloConfig().Http
	require.NotEqual(t, initialHttpConf, afterHttpConf)
	logTestNodeConfig(t, nc)
	t.Log("Logging apollo config")
	apolloCfg = nodecfg.GetApolloConfig()
	logTestNodeConfig(t, &apolloCfg)
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
