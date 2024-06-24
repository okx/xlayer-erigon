package ethconfig

type XLayerConfig struct {
	Apollo ApolloConfig
	Nacos  NacosConfig
}

// NacosConfig is the config for nacos
type NacosConfig struct {
	URLs               string
	NamespaceId        string
	ApplicationName    string
	ExternalListenAddr string
}

// ApolloConfig is the config for apollo
type ApolloConfig struct {
	Enable        bool
	IP            string
	AppID         string
	NamespaceName string
}
