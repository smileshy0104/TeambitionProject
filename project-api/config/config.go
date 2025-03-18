package config

// 导入必要的包
import (
	"github.com/spf13/viper"
	"log"
	"os"
)

// C 是配置的全局实例
var C = InitConfig()

// Config 是应用程序配置的结构体
type Config struct {
	viper      *viper.Viper
	SC         *ServerConfig
	GC         *GrpcConfig
	EtcdConfig *EtcdConfig
}

// ServerConfig 服务器配置的结构体
type ServerConfig struct {
	Name string
	Addr string
}

// GrpcConfig gRPC配置的结构体
type GrpcConfig struct {
	Name string
	Addr string
}

// EtcdConfig Etcd配置的结构体
type EtcdConfig struct {
	Addrs []string
}

// InitConfig 初始化配置并返回配置实例
func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	// 获取当前工作目录
	workDir, _ := os.Getwd()
	// 设置配置文件的名称和类型
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	// 添加配置文件的路径
	conf.viper.AddConfigPath("/etc/ms_project/user")
	conf.viper.AddConfigPath(workDir + "/config")
	// 读取配置文件
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	// 读取服务器和Etcd配置
	conf.ReadServerConfig()
	conf.ReadEtcdConfig()
	return conf
}

// ReadServerConfig 从viper实例中读取服务器配置信息
func (c *Config) ReadServerConfig() {
	sc := &ServerConfig{}
	// 从配置文件中获取服务器的名称和地址
	sc.Name = c.viper.GetString("server.name")
	sc.Addr = c.viper.GetString("server.addr")
	c.SC = sc
}

// ReadEtcdConfig 从viper实例中读取Etcd配置信息
func (c *Config) ReadEtcdConfig() {
	ec := &EtcdConfig{}
	var addrs []string
	// 从配置文件中获取Etcd的地址列表
	err := c.viper.UnmarshalKey("etcd.addrs", &addrs)
	if err != nil {
		log.Fatalln(err)
	}
	ec.Addrs = addrs
	c.EtcdConfig = ec
}
