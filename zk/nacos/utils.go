package nacos

import (
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

func getServerConfigs(urls string) ([]constant.ServerConfig, error) {
	// nolint
	var configs []constant.ServerConfig
	for _, url := range strings.Split(urls, ",") {
		laddr := strings.Split(url, ":")
		serverPort, err := strconv.Atoi(laddr[1])
		if err != nil {
			return nil, err
		}
		configs = append(configs, constant.ServerConfig{
			IpAddr: laddr[0],
			Port:   uint64(serverPort),
		})
	}
	return configs, nil
}
