package config

import "project-common/kafka"

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
