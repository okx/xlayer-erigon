package ethconfig

type XLayerConfig struct {
	Apollo        ApolloClientConfig
	Nacos         NacosConfig
	EnableInnerTx bool
}

// NacosConfig is the config for nacos
type NacosConfig struct {
	URLs               string
	NamespaceId        string
	ApplicationName    string
	ExternalListenAddr string
}

// ApolloClientConfig is the config for apollo
type ApolloClientConfig struct {
	Enable        bool
	IP            string
	AppID         string
	NamespaceName string
}
