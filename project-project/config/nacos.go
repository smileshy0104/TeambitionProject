package config

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"log"
)

// NacosClient 是用于连接 Nacos 并管理配置的客户端。
type NacosClient struct {
	confClient config_client.IConfigClient // confClient 用于与 Nacos 的配置管理功能进行交互。
	group      string                      // group 是 Nacos 中配置的分组名称。
}

// InitNacosClient 初始化并返回一个 NacosClient 实例。
// 此函数读取引导配置，基于该配置创建一个 Nacos 客户端配置，
// 然后使用此配置创建一个新的 NacosClient 实例。
func InitNacosClient() *NacosClient {
	// 初始化引导配置，其中包含 Nacos 的基础配置信息。
	bootConf := InitBootstrap()

	// 创建 Nacos 客户端配置。
	clientConfig := constant.ClientConfig{
		NamespaceId:         bootConf.NacosConfig.Namespace, // 可以通过不同的 namespaceId 创建多个客户端以支持多命名空间。当命名空间为公共时，此处留空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// 创建 Nacos 服务器配置。
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      bootConf.NacosConfig.IpAddr,
			ContextPath: bootConf.NacosConfig.ContextPath,
			Port:        uint64(bootConf.NacosConfig.Port),
			Scheme:      bootConf.NacosConfig.Scheme,
		},
	}

	// 使用客户端和服务器配置创建一个新的 Nacos 配置客户端。
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}

	// 创建并返回一个新的 NacosClient 实例。
	nc := &NacosClient{
		confClient: configClient,
		group:      bootConf.NacosConfig.Group,
	}
	return nc
}
