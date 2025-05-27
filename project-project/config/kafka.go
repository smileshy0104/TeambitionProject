package config

import (
	"context"
	"go.uber.org/zap"
	"project-common/kafka"
	"project-project/internal/dao"
	"project-project/internal/repo"
	"time"
)

// 声明一个全局变量 kw，用于向 Kafka 发送消息
var kw *kafka.KafkaWriter

// InitKafkaWriter 初始化 KafkaWriter
// 该函数负责连接到 Kafka 服务器，并返回一个关闭连接的函数
func InitKafkaWriter() func() {
	// 获取 KafkaWriter 实例，连接到指定的 Kafka 服务器地址
	kw = kafka.GetWriter("localhost:9092")
	// 返回一个函数，当需要关闭 Kafka 连接时可以调用此函数
	return kw.Close
}

// SendLog 向 Kafka 发送日志数据
// 该函数将数据发送到名为 "msproject_log" 的 Kafka 主题
func SendLog(data []byte) {
	// 创建 LogData 实例，指定主题和数据
	kw.Send(kafka.LogData{
		Topic: "msproject_log",
		Data:  data,
	})
}

// SendCache 向 Kafka 发送缓存数据
// 该函数将数据发送到名为 "msproject_cache" 的 Kafka 主题
func SendCache(data []byte) {
	// 创建 LogData 实例，指定主题和数据
	kw.Send(kafka.LogData{
		Topic: "msproject_cache",
		Data:  data,
	})
}

type KafkaCache struct {
	R     *kafka.KafkaReader
	cache repo.Cache
}

func (c *KafkaCache) DeleteCache() {
	for {
		message, err := c.R.R.ReadMessage(context.Background())
		if err != nil {
			zap.L().Error("DeleteCache ReadMessage err", zap.Error(err))
			continue
		}
		zap.L().Info("收到缓存", zap.String("value", string(message.Value)))
		if "task" == string(message.Value) {
			fields, err := c.cache.HKeys(context.Background(), "task")
			if err != nil {
				zap.L().Error("DeleteCache HKeys err", zap.Error(err))
				continue
			}
			time.Sleep(1 * time.Second)
			c.cache.Delete(context.Background(), fields)
		}
	}

}

func NewCacheReader() *KafkaCache {
	reader := kafka.GetReader([]string{"localhost:9092"}, "cache_group", "msproject_cache")
	return &KafkaCache{
		R:     reader,
		cache: dao.Rc,
	}
}
