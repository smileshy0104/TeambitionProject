package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
	"os"
	"project-common/logs"
)

// C 是配置的全局实例
var C = InitConfig()

// Config 是应用配置的结构体
type Config struct {
	viper       *viper.Viper
	SC          *ServerConfig
	GC          *GrpcConfig
	EtcdConfig  *EtcdConfig
	MysqlConfig *MysqlConfig
	JwtConfig   *JwtConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string
	Addr string
}

// GrpcConfig gRPC服务配置
type GrpcConfig struct {
	Name    string
	Addr    string
	Version string
	Weight  int64
}

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Addrs []string
}

// MysqlConfig MySQL数据库配置
type MysqlConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Db       string
}

// JwtConfig JWT配置
type JwtConfig struct {
	AccessExp     int64
	RefreshExp    int64
	AccessSecret  string
	RefreshSecret string
}

// InitConfig 初始化配置
func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath("/etc/ms_project/user")
	conf.viper.AddConfigPath(workDir + "/config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	conf.ReadServerConfig()
	conf.InitZapLog()
	conf.ReadGrpcConfig()
	conf.ReadEtcdConfig()
	conf.InitMysqlConfig()
	conf.InitJwtConfig()
	return conf
}

// InitZapLog 初始化日志
func (c *Config) InitZapLog() {
	lc := &logs.LogConfig{
		DebugFileName: c.viper.GetString("zap.debugFileName"),
		InfoFileName:  c.viper.GetString("zap.infoFileName"),
		WarnFileName:  c.viper.GetString("zap.warnFileName"),
		MaxSize:       c.viper.GetInt("maxSize"),
		MaxAge:        c.viper.GetInt("maxAge"),
		MaxBackups:    c.viper.GetInt("maxBackups"),
	}
	err := logs.InitLogger(lc)
	if err != nil {
		log.Fatalln(err)
	}
}

// ReadServerConfig 读取服务器配置
func (c *Config) ReadServerConfig() {
	sc := &ServerConfig{}
	sc.Name = c.viper.GetString("server.name")
	sc.Addr = c.viper.GetString("server.addr")
	c.SC = sc
}

// ReadGrpcConfig 读取gRPC配置
func (c *Config) ReadGrpcConfig() {
	gc := &GrpcConfig{}
	gc.Name = c.viper.GetString("grpc.name")
	gc.Addr = c.viper.GetString("grpc.addr")
	gc.Version = c.viper.GetString("grpc.version")
	gc.Weight = c.viper.GetInt64("grpc.weight")
	c.GC = gc
}

// ReadRedisConfig 读取Redis配置
func (c *Config) ReadRedisConfig() *redis.Options {
	return &redis.Options{
		Addr:     c.viper.GetString("redis.host") + ":" + c.viper.GetString("redis.port"),
		Password: c.viper.GetString("redis.password"),
		DB:       c.viper.GetInt("redis.db"),
	}
}

// ReadEtcdConfig 读取Etcd配置
func (c *Config) ReadEtcdConfig() {
	ec := &EtcdConfig{}
	var addrs []string
	err := c.viper.UnmarshalKey("etcd.addrs", &addrs)
	if err != nil {
		log.Fatalln(err)
	}
	ec.Addrs = addrs
	c.EtcdConfig = ec
}

// InitMysqlConfig 初始化MySQL配置
func (c *Config) InitMysqlConfig() {
	mc := &MysqlConfig{
		Username: c.viper.GetString("mysql.username"),
		Password: c.viper.GetString("mysql.password"),
		Host:     c.viper.GetString("mysql.host"),
		Port:     c.viper.GetInt("mysql.port"),
		Db:       c.viper.GetString("mysql.db"),
	}
	c.MysqlConfig = mc
}

// InitJwtConfig 初始化JWT配置
func (c *Config) InitJwtConfig() {
	mc := &JwtConfig{
		AccessSecret:  c.viper.GetString("jwt.accessSecret"),
		AccessExp:     c.viper.GetInt64("jwt.accessExp"),
		RefreshExp:    c.viper.GetInt64("jwt.refreshExp"),
		RefreshSecret: c.viper.GetString("jwt.refreshSecret"),
	}
	c.JwtConfig = mc
}
