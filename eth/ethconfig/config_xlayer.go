package ethconfig

// XLayerConfig is the X Layer config used on the eth backend
type XLayerConfig struct {
	Nacos         NacosConfig
	EnableInnerTx bool
}

var DefaultXLayerConfig = &XLayerConfig{}

// NacosConfig is the config for nacos
type NacosConfig struct {
	URLs               string
	NamespaceId        string
	ApplicationName    string
	ExternalListenAddr string
}
