package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

// BootConf 是用于初始化和读取配置文件的结构体
type BootConf struct {
	viper       *viper.Viper
	NacosConfig *NacosConfig
}

// ReadNacosConfig 从 viper 实例中读取 Nacos 配置信息
func (c *BootConf) ReadNacosConfig() {
	nc := &NacosConfig{}
	c.viper.UnmarshalKey("nacos", nc)
	c.NacosConfig = nc
}

// NacosConfig 是 Nacos 服务器的配置信息
type NacosConfig struct {
	Namespace   string
	Group       string
	IpAddr      string
	Port        int
	ContextPath string
	Scheme      string
}

// InitBootstrap 初始化 BootConf 结构体并读取配置文件
func InitBootstrap() *BootConf {
	conf := &BootConf{viper: viper.New()}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("bootstrap")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath(workDir + "/config")
	//conf.viper.AddConfigPath("/config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	conf.ReadNacosConfig()
	return conf
}
